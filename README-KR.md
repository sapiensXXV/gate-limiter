# Gate Limiter

한국어 | [English](./README.md)

[![Golang](https://img.shields.io/badge/Go-1.23.3-00ADD8?style=flat\&logo=Go)](https://go.dev/doc/)
[![NPM](https://img.shields.io/badge/npm-reference-CB3837?style=flat\&logo=npm\&logoColor=CB3837\&labelColor=747474)](https://www.npmjs.com/package/@sapiensxxv/gate-limiter-cli)
![HomeBrew](https://img.shields.io/badge/Homebrew-reference-FBB040?style=flat\&logo=Homebrew\&logoColor=FBB040)
[![Docker](https://img.shields.io/badge/Docker-reference-2496ED?style=flat\&logo=Docker\&logoColor=2496ED)](https://hub.docker.com/repository/docker/sjhn/gate-limiter/general)

---

## 소개 (Introduction)

**gate-limiter**는 API 남용을 방지하고 사용자 간 리소스 사용의 공정성을 보장하기 위해 설계된 **구성형(rate-configurable) 레이트 리미팅(rate-limiting) 미들웨어**입니다.
Go 언어로 작성되었으며, 다음 다섯 가지 레이트 리미팅 알고리즘을 제공합니다:

* Token Bucket
* Leaky Bucket
* Fixed Window Counter
* Sliding Window Log
* Sliding Window Counter

고부하 환경에서도 안정적으로 동작하도록 최적화되어 있으며, 배포가 간편하고 설정이 유연합니다.
Docker를 이용해 독립 실행형 서비스로 운영할 수 있고, RESTful API를 통해 요청 허용 여부를 실시간으로 판단할 수 있습니다.

---

## 설치 (Installation)

```bash
# Homebrew
homebrew install gate-limiter

# NPM
npm install -g @sapiensxxv/gate-limiter-cli

# docker compose
git clone https://github.com/your-org/gate-limiter.git
cd gate-limiter/docker
export GATE_LIMITER_TAG=v0.1.0  # 선택 사항
docker compose up -d

# docker 이미지 단독 실행
docker pull sjhn/gate-limiter:latest
docker run -d \
  -p 8081:8081 \
  -v /path/to/config.yml:/etc/gate-limiter/config.yml:ro \
  -e GATE_LIMITER_CONFIG=/etc/gate-limiter/config.yml \
  --name gate-limiter \
  sjhn/gate-limiter:latest
```

### 주의사항 (Notes)

* **NPM 사용 시**

  * `config.yml` 파일이 명령어 실행 디렉토리에 존재해야 합니다.
  * `gate-limiter` 명령어로 레이트 리미터를 실행합니다.

* **docker compose 사용 시**

  * `config.yml` 파일이 컨테이너 내부에 포함되어 있습니다.
  * `GATE_LIMITER_CONFIG` 환경변수는 `docker-compose.yml`에 이미 정의되어 있습니다.
  * `config.yml`의 `port`를 변경하면, `docker-compose.yml`의 포트 매핑과 `Dockerfile`의 `EXPOSE` 값도 함께 맞춰야 합니다.

* **docker 이미지 단독 사용 시**

  * `config.yml` 파일을 준비하여 컨테이너에 마운트해야 합니다.
  * `GATE_LIMITER_CONFIG` 환경변수는 컨테이너 내부 설정 파일 경로를 가리켜야 합니다.
  * `config.yml`의 `port`를 변경하면, `-p` 포트 매핑과 `EXPOSE` 값도 함께 맞춰야 합니다.

---

## 설정 (Configuration)

### Root 설정 (`RootRateLimiterConfig`)

| Key           | Type                                    | Description      |
| ------------- | --------------------------------------- | ---------------- |
| `rateLimiter` | [RateLimiterConfig](#ratelimiterconfig) | 레이트 리미팅 설정 루트 객체 |
| `redis`       | [RedisClientConfig](#redisclientconfig) | Redis 클라이언트 설정   |

---

### RateLimiterConfig

| Key        | Type                              | Description                                                                                                      |
| ---------- | --------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `strategy` | `string`                          | 알고리즘 선택 (`token_bucket`, `leaky_bucket`, `fixed_window_counter`, `sliding_window_counter`, `sliding_window_log`) |
| `identity` | [ClientIdentity](#clientidentity) | 클라이언트 식별 방식                                                                                                      |
| `client`   | [ClientLimit](#clientlimit)       | 전역 클라이언트 레이트 제한                                                                                                  |
| `apis`     | `[Api]`                           | API 단위 제한 규칙                                                                                                     |
| `target`   | `string`                          | 허용 요청 전달 대상 도메인 URL                                                                                              |
| `port`     | `int`                             | 서버 포트 (기본값: `8081`)                                                                                              |

---

### ClientIdentity

| Key      | Type     | Description                              |
| -------- | -------- | ---------------------------------------- |
| `key`    | `string` | 식별 기준 (`ipv4` 또는 `cookie`)               |
| `header` | `string` | 사용자 식별용 HTTP 헤더 (ipv4일 때만 필요)           |

---

### ClientLimit

| Key             | Type  | Description |
| --------------- | ----- | ----------- |
| `limit`         | `int` | 최대 요청 수     |
| `windowSeconds` | `int` | 시간 윈도우(초)   |

---

### Api

| Key             | Type                                | Description    |
| --------------- | ----------------------------------- | -------------- |
| `identifier`    | `string`                            | API 고유 식별자     |
| `path`          | [RateLimiterPath](#ratelimiterpath) | API 경로 정의      |
| `method`        | `string`                            | HTTP 메서드       |
| `limit`         | `int`                               | 요청 제한 수        |
| `windowSeconds` | `int`                               | 윈도우 시간         |
| `refillSeconds` | `int`                               | 토큰 리필 주기       |
| `expireSeconds` | `int`                               | 저장소 만료 시간      |
| `target`        | `string`                            | API별 전달 대상(옵션) |

---

### RateLimiterPath

| Key          | Type     | Description                 |
| ------------ | -------- | --------------------------- |
| `expression` | `string` | 경로 표현 방식 (`regex`, `plain`) |
| `value`      | `string` | 경로 값                        |

---

### RedisClientConfig

| Key        | Type     | Description  |
| ---------- | -------- | ------------ |
| `host`     | `string` | Redis 서버 주소  |
| `port`     | `int`    | Redis 포트     |
| `password` | `string` | Redis 비밀번호   |
| `db`       | `int`    | Redis DB 인덱스 |

---

## 예시 설정 시나리오 (Example Config Scenarios)

각 설정 예시는 실제 서비스 환경에서 사용 가능한 구조를 기반으로 합니다.

1. 전역 클라이언트 제한 + API별 보호
2. 민감 API 보호 구조 (로그인/결제)
3. 고트래픽 공개 API (Burst 허용)
4. 무거운 API용 큐 기반 트래픽 제어

---

### 1. 전역 클라이언트 제한 + API별 제한

**상황 설명:**
전체 사용자 트래픽은 분당 50회로 제한하면서, 댓글 작성 API는 남용 방지를 위해 분당 5회만 허용하는 구조

```yml
rateLimiter:  
  strategy: sliding_window_counter  
  identity:  
    key: ipv4  
    header: X-Forwarded-For  

  client:
    limit: 50  
    windowSeconds: 60  

  apis:
    - identifier: comment_write  
      path:  
        expression: regex  
        value: ^/api/item/\d+/comment$  
      method: POST  
      limit: 5  
      windowSeconds: 60  
      refillSeconds: 60  
      expireSeconds: 3600  

  target: https://mywebsitedomain.com
  port: 8081  # 기본값

redis:
  host: localhost
  port: 6379
  password:
  db: 0
```

**의미:**

* 클라이언트(IP 기준)는 전역적으로 분당 50회 요청 가능
* 댓글 작성 API는 분당 5회로 추가 제한
* 통과 요청은 백엔드 서비스로 전달

---

### 2. 민감 API 보호 구조

**상황 설명:**
전체 트래픽은 제한하지 않고, 로그인·결제·회원가입 등 민감 API만 보호하는 구조

```yml
rateLimiter:
  strategy: fixed_window_counter
  identity:
    key: cookie

  apis:
    - identifier: login_api
      path:
        expression: plain
        value: /api/auth/login
      method: POST
      limit: 5
      windowSeconds: 60
      expireSeconds: 600

    - identifier: payment_api
      path:
        expression: plain
        value: /api/payment
      method: POST
      limit: 3
      windowSeconds: 60
      expireSeconds: 600

  target: https://api.myservice.com

redis:
  host: redis
  port: 6379
  db: 0
```

**의미:**

* 쿠키 기반 클라이언트 식별
* 전역 제한 없음
* 로그인: 분당 5회 제한
* 결제: 분당 3회 제한
* 민감 API 중심 보호 구조

---

### 3. 고트래픽 공개 API (Token Bucket)

**상황 설명:**
Burst 트래픽을 허용하면서 평균 처리량을 제한하는 구조 (공개 API, 외부 클라이언트)

```yml
rateLimiter:
  strategy: token_bucket
  identity:
    key: ipv4
    header: X-Forwarded-For

  apis:
    - identifier: public_api
      path:
        expression: plain
        value: /api/public/data
      method: GET
      limit: 100
      refillSeconds: 10
      windowSeconds: 10
      expireSeconds: 600

  target: https://public-api.service.com

redis:
  host: localhost
  port: 6379
  db: 0
```

**의미:**

* 순간 Burst 트래픽 허용
* 평균 처리량 제한 유지
* 공개 API에 적합한 구조

---

### 4. 큐 기반 트래픽 제어 (Leaky Bucket)

**상황 설명:**
요청을 버리지 않고, 처리 속도만 강제로 제한하는 구조 (ML API, 고부하 처리 API)

```yml
rateLimiter:
  strategy: leaky_bucket
  identity:
    key: cookie

  apis:
    - identifier: heavy_api
      path:
        expression: plain
        value: /api/ml/process
      method: POST
      limit: 20
      windowSeconds: 1
      expireSeconds: 600

  target: https://ml-backend.service.com

redis:
  host: localhost
  port: 6379
  db: 0
```

**의미:**

* 요청 큐잉 처리
* 고정 처리 속도
* 트래픽 스무딩
* 고부하 API 안정화

---

## 알고리즘 (Algorithms)

Gate Limiter는 다음 알고리즘을 지원합니다:

* Token Bucket
* Leaky Bucket
* Fixed Window Counter
* Sliding Window Log
* Sliding Window Counter

설정 경로:

```yml
rateLimiter.strategy
```

### Token Bucket
요청마다 토큰을 소비하며, 토큰이 존재하면 요청을 허용하고 없으면 거부합니다.  
토큰은 일정 주기마다 자동으로 보충됩니다.

Parameters:
* Bucket size → `rateLimiter.apis.limit`  
* Refill interval → `rateLimiter.apis.refillSeconds`

### Leaky Bucket
고정 처리 속도로 요청을 큐(Go channel 모델) 기반으로 처리합니다.  
큐가 가득 차면 요청은 드롭됩니다.

Parameters:
* Queue size → `rateLimiter.apis.limit`  
* Processing interval → `rateLimiter.apis.windowSeconds`

### Fixed Window Counter
고정된 시간 윈도우 단위로 요청 수를 카운팅합니다.

Rules:
* 요청마다 카운터 증가  
* counter ≥ limit → 요청 거부  
* counter < limit → 요청 허용  

Parameters:
* Window size → `rateLimiter.apis.limit`  
* Window duration → `rateLimiter.apis.windowSeconds`

### Sliding Window Log
요청 타임스탬프를 저장하고, 이동하는 시간 윈도우 내 요청 수를 기준으로 판단합니다.

Rules:
* 만료된 타임스탬프 제거  
* count < limit → 요청 허용  
* count ≥ limit → 요청 거부  

Parameters:
* Window size → `rateLimiter.apis.limit`  
* Window duration → `rateLimiter.apis.windowSeconds`

### Sliding Window Counter
고정 윈도우와 슬라이딩 윈도우를 결합한 방식으로,  
이전 윈도우와 현재 윈도우 간 가중치 기반 근사 계산(weighted approximation)을 사용합니다.

Parameters:
* Window size → `rateLimiter.apis.limit`  
* Window duration → `rateLimiter.apis.windowSeconds`

## 작성자 (Author)

* Jaehoon So
* Email: [jhspacelover@naver.com](mailto:jhspacelover@naver.com)

---

## 라이선스 (License)

`gate-limiter`는 **MIT License**로 배포됩니다.
자세한 내용은 [LICENSE](https://github.com/sapiensXXV/gate-limiter/blob/main/LICENSE) 파일을 참고하세요.
