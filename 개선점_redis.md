# Gate Limiter — Redis 사용 개선점 분석

## 1. 원자성(Atomicity) 문제

### 1-1. SlidingWindowLog: 요청 등록 후 제한 검사 (순서 오류 + 비원자적)

**파일:** `internal/limiter/strategy/sliding_window_log_limiter.go:63-80`

```go
// 1. 오래된 항목 삭제
err = l.RedisClient.RemoveOldEntries(key, ...)
// 2. 새 요청을 먼저 추가 ← 문제!
err = l.RedisClient.AddToSortedSet(key, now.String(), now)
// 3. 그 후에 카운트 검사
size := l.RedisClient.ZSetSize(key)
if size > api.Limit { return denied }
```

**문제점:**
- 거부될 요청이라도 sorted set에 먼저 추가되어, 정상 요청의 슬롯을 점유
- 3개의 Redis 명령이 별도로 실행되므로 동시 요청 간 Race Condition 발생
- 예: limit=5일 때 동시에 10개 요청이 오면, 모두 `AddToSortedSet` 후 `ZSetSize`를 확인하므로 5개 이상 통과 가능

**개선 방향:** Lua 스크립트로 `ZREMRANGEBYSCORE → ZCARD → (limit 미만이면) ZADD`를 원자적으로 처리

```lua
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, '0', tostring(now - window))
local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, now, member)
    redis.call('EXPIRE', key, window)
    return {1, limit - count - 1}  -- allowed, remaining
else
    return {0, 0}  -- denied
end
```

---

### 1-2. SlidingWindowCounter: 비원자적 다중 Redis 호출

**파일:** `internal/limiter/strategy/sliding_window_counter_limiter.go:58-113`

```go
l.RedisClient.RemoveOldEntries(...)    // 호출 1
l.RedisClient.ZRemRangeByScore(...)    // 호출 2 (호출 1과 중복)
l.RedisClient.AddToSortedSet(...)      // 호출 3
l.RedisClient.ZSetSize(key)            // 호출 4
l.RedisClient.ZCount(key, ...)         // 호출 5
```

**문제점:**
- 5번의 별도 Redis 왕복이 발생 (네트워크 지연 × 5)
- `RemoveOldEntries`와 `ZRemRangeByScore`가 거의 동일한 범위를 삭제 (중복 호출)
- 모든 연산이 비원자적이므로 동시 요청 시 정확한 카운트 보장 불가

**개선 방향:** 단일 Lua 스크립트로 전체 로직 통합, 중복 삭제 제거

---

## 2. 인터페이스 설계 문제

### 2-1. `RedisClient` 인터페이스가 과도하게 비대함

**파일:** `internal/limiter/types/redis.go`

현재 `RedisClient` 인터페이스는 18개의 메서드를 가지고 있지만, 각 전략은 그 중 일부만 사용합니다:

| 전략 | 사용하는 메서드 |
|------|----------------|
| TokenBucket | `Eval` |
| FixedWindowCounter | `Eval`, `Expire` |
| SlidingWindowLog | `RemoveOldEntries`, `AddToSortedSet`, `ZSetSize`, `GetOldestEntry` |
| SlidingWindowCounter | `RemoveOldEntries`, `ZRemRangeByScore`, `AddToSortedSet`, `ZSetSize`, `ZCount` |
| LeakyBucket | Redis 미사용 |

**문제점:**
- Interface Segregation Principle 위반 — Mock이나 대체 구현 작성 시 사용하지 않는 메서드까지 모두 구현해야 함
- `HGetObject`, `HSetObject`는 미구현 상태 (TODO/stub)
- `GetObject`는 `TokenBucket` 타입에 하드코딩되어 범용성 없음 (현재는 Lua로 대체되어 미사용)

**개선 방향:** 각 전략이 필요로 하는 최소 인터페이스로 분리하거나, 모든 전략을 Lua 스크립트 기반으로 전환하여 `Eval` 메서드만 사용하도록 단순화

```go
// 이상적인 형태: 모든 전략이 Lua로 전환되면
type RedisClient interface {
    Eval(script string, keys []string, args ...interface{}) (interface{}, error)
    Ping() error
}
```

---

### 2-2. `GetObject`가 `TokenBucket` 타입에 하드코딩

**파일:** `pkg/redisclient/client.go:55-71`

```go
func (d *DefaultRedisClient) GetObject(key string) (interface{}, error) {
    // ...
    var bucket types.TokenBucket           // ← 하드코딩
    err = json.Unmarshal(val, &bucket)
    return &bucket, nil
}
```

`GetObject`라는 범용적인 이름이지만 내부에서 `types.TokenBucket`으로만 역직렬화합니다. 현재는 TokenBucket이 Lua 스크립트 기반 Hash로 전환되어 이 메서드가 사용되지 않지만, 범용 Redis 클라이언트로서 설계가 잘못되어 있습니다.

**개선 방향:** 제네릭 또는 `interface{}` 역직렬화 대상을 인자로 받는 방식으로 변경하거나, 미사용 메서드이므로 제거

---

## 3. Context 전파 미비

### 3-1. 요청별 Context가 Redis 호출에 전달되지 않음

**파일:** `pkg/redisclient/client.go:18-20`

```go
type DefaultRedisClient struct {
    ctx    context.Context        // context.Background() 고정
    client *redis.Client
}
```

`DefaultRedisClient`는 생성 시 `context.Background()`를 저장하고, 모든 Redis 호출에 이 고정 context를 사용합니다. HTTP 요청의 context가 전파되지 않으므로:

- 클라이언트가 요청을 취소해도 Redis 작업은 계속 실행됨
- 요청 타임아웃이 Redis 호출에 적용되지 않음
- LeakyBucket에서 `queuedRequest.Request.Context().Done()`으로 취소를 감지하지만, 이후 Redis 호출이 있다면 취소가 전파되지 않음

**개선 방향:** `RedisClient` 인터페이스의 각 메서드에 `context.Context` 매개변수를 추가하거나, 최소한 `WithContext(ctx)` 메서드를 제공

```go
type RedisClient interface {
    Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
}
```

---

## 4. 연결 관리 / 운영

### 4-1. Redis 연결 풀 설정 미비

**파일:** `pkg/redisclient/client.go:31-35`

```go
rc.client = redis.NewClient(&redis.Options{
    Addr:     config.Host + ":" + strconv.Itoa(config.Port),
    Password: config.Password,
    DB:       config.DB,
})
```

`redis.Options`에서 다음 설정이 누락되어 있어, 고부하 환경에서 성능 문제가 발생할 수 있습니다:

| 설정 | 기본값 | 권장 |
|------|--------|------|
| `PoolSize` | CPU수 × 10 | 워크로드에 맞게 조정 |
| `MinIdleConns` | 0 | cold start 방지를 위해 설정 |
| `MaxRetries` | 0 | 일시적 네트워크 오류 대응 |
| `DialTimeout` | 5s | 운영 환경에 맞게 조정 |
| `ReadTimeout` | 3s | Lua 스크립트 실행 시간 고려 |
| `WriteTimeout` | 3s | - |
| `PoolTimeout` | ReadTimeout + 1s | - |

**개선 방향:** `config.yml`에서 풀 설정을 받거나, 합리적인 기본값을 설정

---

### 4-2. Redis 연결 실패 시 `log.Fatal`로 프로세스 종료

**파일:** `pkg/redisclient/client.go:37-39`

```go
if err := rc.client.Ping(rc.ctx).Err(); err != nil {
    log.Fatal("redis client connection fail")
}
```

초기 연결 실패 시 `log.Fatal`로 즉시 종료되므로:
- 에러 메시지에 원인이 포함되지 않음 (`err`를 출력하지 않음)
- 재시도 로직 없음 — Redis가 수 초 후 준비될 수 있는 상황에서도 즉시 포기
- 테스트에서 `log.Fatal`이 호출되면 테스트 프로세스 전체가 종료됨

**개선 방향:** 에러를 반환하여 호출자가 처리하도록 변경, 재시도/백오프 로직 추가

---

### 4-3. TLS 연결 미지원

Redis 연결이 평문 TCP만 지원합니다. 클라우드 환경(AWS ElastiCache, GCP Memorystore 등)에서는 TLS가 필수인 경우가 많습니다.

**개선 방향:** `config.yml`에 `tls: true` 옵션 추가, `redis.Options.TLSConfig` 설정

---

### 4-4. Redis Sentinel / Cluster 미지원

현재는 단일 Redis 인스턴스만 지원합니다. 프로덕션 환경에서 고가용성을 위해 Sentinel 또는 Cluster 모드 지원이 필요합니다.

**개선 방향:** `redis.NewFailoverClient` (Sentinel) 또는 `redis.NewClusterClient` 지원을 config 옵션으로 제공

---

## 5. 에러 처리

### 5-1. `ZSetSize`가 에러를 삼킴

**파일:** `pkg/redisclient/client.go:110-116`

```go
func (d *DefaultRedisClient) ZSetSize(key string) int {
    size, err := d.client.ZCard(d.ctx, key).Result()
    if err != nil {
        log.Println("redisclient: get zset size fail")
        // err 무시, 0 반환
    }
    return int(size)
}
```

에러 발생 시 0을 반환하면 실제 카운트가 0인 것처럼 보여, 제한이 적용되지 않을 수 있습니다.

### 5-2. `GetOldestEntry`가 빈 결과에서 패닉

**파일:** `pkg/redisclient/client.go:118-124`

```go
func (d *DefaultRedisClient) GetOldestEntry(key string) (redis.Z, error) {
    vals, err := d.client.ZRangeWithScores(d.ctx, key, 0, 0).Result()
    if err != nil {
        log.Println("redisclient: get oldest entry fail")
    }
    return vals[0], err  // vals가 비어있으면 index out of range 패닉
}
```

### 5-3. `Expire` 메서드가 에러를 반환하지 않음

**파일:** `pkg/redisclient/client.go:137-139`

```go
func (d *DefaultRedisClient) Expire(key string, seconds int) {
    d.client.Expire(d.ctx, key, time.Duration(seconds)*time.Second)
    // 에러 무시
}
```

TTL 설정 실패 시 키가 영구적으로 남아 메모리 누수를 유발할 수 있습니다.

---

## 6. Lua 스크립트 최적화

### 6-1. EVALSHA 미사용

현재 모든 Lua 스크립트 실행이 `EVAL`로 수행됩니다. 매번 스크립트 전문이 Redis로 전송됩니다.

**개선 방향:** `go-redis`의 `redis.NewScript()`를 사용하면 스크립트가 자동으로 SHA 기반(`EVALSHA`)으로 캐싱되어 네트워크 전송량이 줄어듭니다.

```go
var tokenBucketScript = redis.NewScript(`...lua script...`)

// 사용 시
result, err := tokenBucketScript.Run(ctx, client, keys, args...).Result()
```

---

## 7. 키 설계

### 7-1. 네임스페이스 미적용

현재 Redis 키 형식: `{strategy}:{ip}:{identifier}` (예: `token_bucket:127.0.0.1:comment_write`)

같은 Redis 인스턴스를 다른 애플리케이션과 공유하는 경우 키 충돌이 발생할 수 있습니다.

**개선 방향:** 글로벌 프리픽스 추가: `gate-limiter:{strategy}:{ip}:{identifier}`

### 7-2. 클라이언트 제한 키와 API 제한 키의 형식 불일치

- API 제한: `{strategy}:{ip}:{identifier}`
- 클라이언트 제한: `client:{ip}:{windowStart}`

키 생성 규칙이 일관되지 않아 디버깅이나 Redis 모니터링 시 혼란을 줄 수 있습니다.

---

## 요약 우선순위

| 우선순위 | 항목 | 영향도 |
|---------|------|--------|
| 높음 | SlidingWindowLog/Counter 원자성 (1-1, 1-2) | 동시 요청 시 Rate Limit 우회 가능 |
| 높음 | Context 전파 (3-1) | 요청 취소가 Redis에 전파되지 않음 |
| 높음 | GetOldestEntry 패닉 (5-2) | 서버 크래시 |
| 중간 | 연결 풀 설정 (4-1) | 고부하 시 성능 저하 |
| 중간 | 에러 처리 개선 (5-1, 5-3) | 조용한 실패로 제한 미적용 |
| 중간 | RedisClient 인터페이스 분리 (2-1) | 유지보수성 |
| 중간 | EVALSHA 캐싱 (6-1) | 네트워크 효율 |
| 낮음 | TLS / Sentinel / Cluster (4-3, 4-4) | 프로덕션 환경 요구사항 |
| 낮음 | 키 네임스페이스 (7-1, 7-2) | 다중 앱 환경 대비 |
