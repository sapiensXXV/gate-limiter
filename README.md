
<h1 align="center">Gate Limiter</h1>

한국어 | [English](./README_EN.md)

[![Golang](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat&logo=Go)](https://go.dev/doc/)
[![NPM](https://img.shields.io/badge/npm-reference-CB3837?style=flat&logo=npm&logoColor=CB3837&labelColor=747474
)](https://www.npmjs.com/package/@sapiensxxv/gate-limiter-cli)
![HomeBrew](https://img.shields.io/badge/Homebrew-reference-FBB040?style=flat&logo=Homebrew&logoColor=FBB040
)
[![Docker](https://img.shields.io/badge/Docker-reference-2496ED?style=flat&logo=Docker&logoColor=2496ED
)](https://hub.docker.com/repository/docker/sjhn/gate-limiter/general)

## 소개
**gate-limiter**는 API 남용을 방지하고 사용자 간 공정한 리소스 사용을 보장하기 위해 설계된, 설정 가능한 요청 처리량 제한(rate limiting) 미들웨어 입니다. Go 언어로 작성되었으며 다음 다섯가지의 처리량제한 알고리즘을 제공합니다.
- 토큰 버킷(Token Bucket)
- 누출 버킷(Leaky Bucket)
- 고정 윈도우 카운터(Fixed Window Counter)
- 슬라이딩 윈도우 로그(Sliding Window Log)
- 슬라이딩 윈도우 카운터(Sliding Window Counter)

배포가 간편하고 설정이 유연하며, 고부하 환경에서도 안정적으로 동작하도록 최적화되어 있습니다. Docker를 이용해 독립 실행형 서비스로 운영할 수 있으며, RESTful API를 통해 요청 허용 여부를 실시간으로 판단할 수 있습니다.

## 설치
```bash
# Homebrew
homebrew install gate-limiter

# NPM
npm install -g @sapiensxxv/gate-limiter-cli

# with docker compose
git clone https://github.com/your-org/gate-limiter.git
cd gate-limiter/docker
export GATE_LIMITER_TAG=v0.1.0  # 또는 생략 가능
docker compose up -d

# only docker image
docker pull sjhn/gate-limiter:latest
docker run -d \
  -p 8081:8081 \
  -v /path/to/config.yml:/etc/gate-limiter/config.yml:ro \
  -e GATE_LIMITER_CONFIG=/etc/gate-limiter/config.yml \
  --name gate-limiter \
  sjhn/gate-limiter:latest
```
- NPM을 활용하는 경우
    - 설정 파일 `config.yml`은 명령어를 실행하는 현재 디렉토리에 존재해야합니다.
    - `gate-limiter` 커맨드로 처리율 제한기를 실행합니다.
- docker compose를 활용하는 경우 
	- 설정 파일 `config.yml`이 컨테이너 내부에 포함되어 있습니다.
	- `GATE_LIMITER_CONFIG` 환경변수는 이미 `docker-compose.yml`에 작성되어 있습니다.
- docker image 를 활용하는 경우
	- 반드시 설정 파일 `config.yml`을 준비하고 해당 경로를 Docker 컨테이너에 마운트 해야 합니다.
	- `GATE_LIMITER_CONFIG` 환경변수는 `config.yml` 파일의 컨테이너 내 경로를 가리켜야합니다.

## 설정 정보
### Root Configuration (`RootRateLimiterConfig`)

| Key           | Type                                    | Description                    |
|---------------|-----------------------------------------|--------------------------------|
| `rateLimiter` | [RateLimiterConfig](#ratelimiterconfig) | 처리율 제한(Throttling) 관련 설정 루트 객체 |
| `redis`       | [RedisClientConfig](#redisclientconfig) | Redis 클라이언트 설정                 |

### RateLimiterConfig

| Key        | Type                              | Description                                                                                                              |
|------------|-----------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| `strategy` | `string`                          | 처리량 제한 알고리즘 선택. (`token_bucket`, `leaky_bucket`, `fixed_window_counter`, `sliding_window_counter`, `sliding_window_log`) |
| `identity` | [ClientIdentity](#clientidentity) | 요청자의 식별 방법 설정                                                                                                            |
| `client`   | [ClientLimit](#clientlimit)       | 클라이언트 단위의 전체 요청 제한                                                                                                       |
| `apis`     | `[Api]`                           | 특정 API 단위 요청 제한 규칙 목록                                                                                                    |
| `target`   | `string`                          | 허용된 요청이 전달될 대상 도메인 URL                                                                                                   |

### ClientIdentity
| Key      | Type     | Description                     |
|----------|----------|---------------------------------|
| `key`    | `string` | 사용자 식별 기준 (예: `ipv4`, `header`) |
| `header` | `string` | 사용자 식별 정보가 담긴 HTTP 헤더 이름        |

### ClientLimit
| Key             | Type  | Description         |
|-----------------|-------|---------------------|
| `limit`         | `int` | 허용 요청 수 한계치         |
| `windowSeconds` | `int` | 요청 제한 시간 윈도우 (초 단위) |

### Api
| Key             | Type                                | Description                        |
|-----------------|-------------------------------------|------------------------------------|
| `identifier`    | `string`                            | API 식별자 (유일해야 함)                   |
| `path`          | [RateLimiterPath](#ratelimiterpath) | API 경로 정의                          |
| `method`        | `string`                            | HTTP 메서드 (GET, POST 등)             |
| `limit`         | `int`                               | 해당 API의 요청 한도                      |
| `windowSeconds` | `int`                               | 윈도우 시간 단위                          |
| `refillSeconds` | `int`                               | 버킷 알고리즘의 토큰 리필 주기                  |
| `expireSeconds` | `int`                               | 메모리/Redis에서 윈도우 또는 버킷 유지 시간 (초 단위) |
| `target`        | `string`                            | 해당 API 호출이 전달될 도메인 (옵션)            |

### RateLimiterPath
| Key          | Type     | Description                                            |
|--------------|----------|--------------------------------------------------------|
| `expression` | `string` | 경로 표현 방식 (`regex`, `plain`)                            |
| `value`      | `string` | API 경로 값. expression이 `regex`이면 정규식, `plain`이면 문자열로 지정 |

### RedisClientConfig
| Key        | Type     | Description     |
|------------|----------|-----------------|
| `host`     | `string` | Redis 서버 주소     |
| `port`     | `int`    | Redis 포트 번호     |
| `password` | `string` | Redis 인증 비밀번호   |
| `db`       | `int`    | Redis DB 인덱스 번호 |

### 설정파일 예시
설정파일 config.yml 예시

```yml
rateLimiter:  
  strategy: sliding_window_counter  
  identity:  
    key: ipv4  
    header: X-Forwarded-For  
  client: # 클라이언트의 전체 처리량 제한  
    limit: 50  
    windowSeconds: 60  
  apis: # 특정 API 처리량 제한  
    - identifier: comment_write  
      path:  
        expression: regex  
        value: ^/api/item/\d+/comment$  
      method: POST  
      limit: 5  
      windowSeconds: 60  
      refillSeconds: 60 #// 토큰 버킷 알고리즘의 경우 토큰 리필 시간  
      expireSeconds: 3600  
  target: https://mywebsitedomain.com # 통과된 요청이 전달될 도메인
redis:
  host: localhost
  port: 6379
  password:
  db: 0
```

## 알고리즘
`gate-limiter` 에서는 아래 다섯가지 알고리즘을 제공합니다.
- 토큰 버킷(Token Bucket)
- 누출 버킷(Leaky Bucket)
- 고정 윈도우 카운터(Fixed Window Counter)
- 슬라이딩 윈도우 로그(Sliding Window Log)
- 슬라이딩 윈도우 카운터(Sliding Window Counter)

알고리즘은 설정 파일 `config.yml` 에서 `rateLimiter.strategy` 필드로 설정할 수 있습니다.
### 토큰 버킷 (Token Bucket)
요청 단위마다 버킷의 토큰을 소비하는 알고리즘입니다. 토큰이 남아있다면 요청이 통과되고, 남아있지 않다면 거부됩니다. 토큰은 주기적으로 채워집니다.
<p align="center">
	<img width="900" alt="스크린샷 2025-08-01 오전 3 10 10" src="https://github.com/user-attachments/assets/de6bd04f-9148-4e0f-98d2-60eb393fb75d" />
</p>

두 가지 파라미터를 조절해야 합니다.
- 버킷 크기: `config.yml`의 `rateLimiter.apis.limit` 값으로 조절할 수 있습니다.
- 토큰 공급 주기: `config.yml`의 `rateLimiter.apis.refillSeconds` 값으로 조절할 수 있습니다.
### 누출 버킷 (Leaky Bucket)
시간 단위로 요청 처리율이 고정되어 있는 알고리즘 입니다. Golang의 채널(channel)을 응용하거 구현되어 있습니다. 요청이 도착하면 채널이 가득차있는지 확인합니다. 채널에 빈자리가 있다면 채널에 요청이 추가되고, 빈자리가 없다면 요청은 버려집니다. 지정된 주기마다 큐에서 요청을 꺼내 처리합니다.
<p align="center">
	<img width="900" alt="스크린샷 2025-08-01 오전 3 13 23" src="https://github.com/user-attachments/assets/62eaa706-97d0-48b1-bfc5-eae9ef80a902" />
</p>

두 가지 파라미터를 조절해야 합니다.

- 큐(채널)의 크기: `config.yml`의 `rateLimiter.apis.limit` 값으로 조절할 수 있습니다.
- 요청 처리 주기: `config.yml`의 `rateLimiter.apis.windowSeconds` 값으로 조절할 수 있습니다.

### 고정 윈도우 카운터 (Fixed Window Counter)
타임라인을 윈도우라는 고정된 단위로 나누고 윈도우마다 카운터를 붙이는 방법입니다.

- 요청이 들어올 때마다 윈도우의 카운터 값이 1증가합니다.
- 윈도우의 카운터 값이 임계치와 같거나 큰 경우 들어오는 요청은 버려집니다.
- 윈도우의 카운터 값이 임계치보다 작은 경우 요청이 받아 들여집니다.

아래의 그림은 1분간 3번의 요청으로 제한된 경우를 나타낸 것입니다.
<p align="center">
	<img width="900" alt="스크린샷 2025-08-01 오전 4 20 58" src="https://github.com/user-attachments/assets/098a5d02-880d-4b84-b4e7-24c5d34a2f0a" />
</p>

두 가지 파라미터를 조절해야 합니다.
- 윈도우 사이즈: `config.yml`의 `rateLimiter.apis.limit` 값으로 조절할 수 있습니다.
- 윈도우 시간 단위: `config.yml`의 `rateLimiter.apis.windowSeconds` 값으로 조절할 수 있습니다.
### 슬라이딩 윈도우 로깅 (Sliding Window Log)
슬라이딩 윈도우 로깅(Sliding Window Logging) 알고리즘은 시간 기반의 요청 제한을 구현하는 방식 중 하나로, 지정된 시간 범위 내의 실제 요청 시각(타임스탬프)을 로그 형태로 저장하여 요청 허용 여부를 판단합니다.
- 로그가 비어있다면 요청을 허용하고 타임스탬프를 기록합니다.
- 로그가 비어있지 않다면
	- 윈도우 범위 밖에 있는 타임스탬프가 있는지 확인하고 삭제합니다.
	- 윈도우 내 타임스탬프의 갯수가 임계치보다 작다면 요청을 허용한다.
	- 윈도우 내 타임스탬프의 갯수가 임계치와 같거나 크다면 요청을 거부한다.

<p align="center">
<img width="900" alt="스크린샷 2025-08-01 오전 4 57 03" src="https://github.com/user-attachments/assets/85fc0c83-b11d-43a2-b148-4104781936e1" />
</p>

두 가지 파라미터를 조절해야 합니다.

- 윈도우 사이즈: `config.yml`의 `rateLimiter.apis.limit` 값으로 조절할 수 있습니다.
- 윈도우 시간 단위: `config.yml`의 `rateLimiter.apis.windowSeconds` 값으로 조절할 수 있습니다.

### 슬라이딩 윈도우 카운터 (Sliding Window Counter)
슬라이딩 윈도우 카운터(Sliding Window Counter) 알고리즘은 고정 윈도우 알고리즘과 이동 윈도우 알고리즘을 결합한 알고리즘입니다. 현재 윈도우가 직전 고정 시간대와 현재 고정 시간대를 차지하고 있는 비율에 따라서 현재 윈도우의 요청 수를 근사치로 계산하는 방법입니다.

<p align="center">
<img width="900" alt="스크린샷 2025-08-01 오전 5 09 27" src="https://github.com/user-attachments/assets/fbe86474-8de0-47ee-bd16-554a7e358d80" />
</p>
 
두 가지 파라미터를 조절해야 합니다.
- 윈도우 사이즈: `config.yml`의 `rateLimiter.apis.limit` 값으로 조절할 수 있습니다.
- 윈도우 시간 단위: `config.yml`의 `rateLimiter.apis.windowSeconds` 값으로 조절할 수 있습니다.

## Author
- [Jaehoon So](https://github.com/sapiensXXV)
- Email: jhspacelover@naver.com
## License
`gate-limiter` is available under the `MIT license`. See the [LICENSE](https://github.com/sapiensXXV/gate-limiter/blob/main/LICENSE) file for more info.
