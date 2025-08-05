
<h1 align="center">Gate Limiter</h1>

한국어 | [English](./README_EN.md)

[![Golang](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat&logo=Go)](https://go.dev/doc/)
[![NPM](https://img.shields.io/badge/npm-reference-CB3837?style=flat&logo=npm&logoColor=CB3837&labelColor=747474
)](https://www.npmjs.com/package/@sapiensxxv/gate-limiter-cli)
![HomeBrew](https://img.shields.io/badge/Homebrew-reference-FBB040?style=flat&logo=Homebrew&logoColor=FBB040
)
[![Docker](https://img.shields.io/badge/Docker-reference-2496ED?style=flat&logo=Docker&logoColor=2496ED
)](https://hub.docker.com/repository/docker/sjhn/gate-limiter/general)

## Introduction
**gate-limiter**는 API 남용을 방지하고 사용자 간 공정한 리소스 사용을 보장하기 위해 설계된, 설정 가능한 요청 처리량 제한(rate limiting) 미들웨어 입니다. Go 언어로 작성되었으며 다음 다섯가지의 처리량제한 알고리즘을 제공합니다.
- 토큰 버킷(Token Bucket)
- 누출 버킷(Leaky Bucket)
- 고정 윈도우 카운터(Fixed Window Counter)
- 슬라이딩 윈도우 로그(Sliding Window Log)
- 슬라이딩 윈도우 카운터(Sliding Window Counter)

배포가 간편하고 설정이 유연하며, 고부하 환경에서도 안정적으로 동작하도록 최적화되어 있습니다. Docker를 이용해 독립 실행형 서비스로 운영할 수 있으며, RESTful API를 통해 요청 허용 여부를 실시간으로 판단할 수 있습니다.
## Installation
```bash
# Homebrew
homebrew install gate-limiter

# NPM
npm install -g @sapiensxxv/gate-limiter

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

- docker compose를 활용하는 경우 
	- 설정 파일 `config.yml`이 컨테이너 내부에 포함되어 있습니다.
	- `GATE_LIMITER_CONFIG` 환경변수는 이미 `docker-compose.yml`에 작성되어 있습니다.
- docker image 를 활용하는 경우
	- 반드시 설정 파일 `config.yml`을 준비하고 해당 경로를 Docker 컨테이너에 마운트 해야 합니다.
	- `GATE_LIMITER_CONFIG` 환경변수는 `config.yml` 파일의 컨테이너 내 경로를 가리켜야합니다.

## Setting
설정파일 config.yml 예시는 아래와 같습니다.

```yml
rateLimiter:  
  strategy: sliding_window_counter  
  # token bucket, leaky bucket, fixed window counter, sliding window counter, sliding_window_log  
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
```

- **rateLimiter**: 설정의 루트. 처리율 제한의 모든 설정정보는 이곳에서 시작한다.
	- **strategy**: 처리량 제한에서 사용할 알고리즘을 선택하는 옵션입니다.
		- `token_bucket`: 토큰 버킷 알고리즘
		- `leaky_bucket`: 누출 버킷 알고리즘
		- `fixed_window_counter`: 고정 윈도우 카운터
		- `sliding_window_log`: 이동 윈도우 로그
		- `sliding_window_counter`: 이동 윈도우 카운터
	- **identity**: 사용자를 어떻게 시별할지 결정하는 옵션입니다.
		- **key**: 사용자를 식별하는 기준 
			- `ipv4`: IPv4 주소를 기준으로 사용자를 식별합니다.
		- **header**: 사용자 식별 정보를 얻어올 수 있는 헤더 이름
	- **apis**: 특정 API의 처리량을 제한하는 옵션입니다. 리스트로 여러가지 API를 명시할 수 있습니다.
		- **key**: api 식별자. 어떠한 문자열이라도 괜찮습니다. 단, 다른 API에 대해 유일해야합니다.
		- **path**: API 경로를 표현하는 옵션입니다.
			- **expression**: API 표현 방식. 이 값에 따라 value 옵션의 해석 방법이 결정됩니다.
				- `regex`: 정규식 표현
				- `plain`: 일반 텍스트 표현
			- **value**: API 경로 표현. expression이 regex 였다면 정규식을, plain이였다면 경로 문자열을 그대로 작성하면 됩니다.
		- **method**: HTTP 메서드
		- **limit**: 요청 임계치. 윈도우나 버킷의 최대 사이즈가 이 옵션에서 결정됩니다.
		- **windowSeconds**: 윈도우 시간 단위. 윈도우 관련 알고리즘을 사용하는 경우 설정해야 합니다.
			- 고정 윈도우 카운터 알고리즘
			- 이동 윈도우 로깅 알고리즘
			- 이동 윈도우 카운터 알고리즘
		- **refillSeconds**: 버킷에 토큰이 채워지는 시간단위. 버킷 관련 알고리즘을 사용하는 경우 설정해야 합니다.
			- 토큰 버킷 알고리즘
			- 누출 버킷 알고리즘
		- **expireSeconds**: 사용되지 않는 버킷이나 윈도우가 메모리/Redis에서 유지되는 시간
	- **target**: 허용된 요청이 전달될 도메인 주소

## Algorithm
gate-limiter 에서는 아래 다섯가지 알고리즘을 제공합니다.
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

## More Info
어떤 알고리즘을 선택해야할지 고민된다면 [처리율 제한기 개발](https://sapiensxxv.github.io/posts/%EC%B2%98%EB%A6%AC%EC%9C%A8-%EC%A0%9C%ED%95%9C%EA%B8%B0-%EA%B0%9C%EB%B0%9C/) 포스트를 참고해주세요

## Author
- [Jaehoon So](https://github.com/sapiensXXV)
- Email: jhspacelover@naver.com
## License
`gate-limmiter` is available under the `MIT license`. See the [LICENSE](https://github.com/sapiensXXV/gate-limiter/blob/main/LICENSE) file for more info.
