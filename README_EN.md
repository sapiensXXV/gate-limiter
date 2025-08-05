<h1 align="center">Gate Limiter</h1>

[한국어](./README.md) | English

[![Golang](https://img.shields.io/badge/Go-1.24.5-00ADD8?style=flat&logo=Go)](https://go.dev/doc/)
[![NPM](https://img.shields.io/badge/npm-reference-CB3837?style=flat&logo=npm&logoColor=CB3837&labelColor=747474
)](https://www.npmjs.com/package/@sapiensxxv/gate-limiter-cli)
![HomeBrew](https://img.shields.io/badge/Homebrew-reference-FBB040?style=flat&logo=Homebrew&logoColor=FBB040
)
[![Docker](https://img.shields.io/badge/Docker-reference-2496ED?style=flat&logo=Docker&logoColor=2496ED
)](https://hub.docker.com/repository/docker/sjhn/gate-limiter/general)

## Introduction
**gate-limiter** is a configurable rate limiting middleware designed to prevent API abuse and ensure fair resource usage among users. Written in Go, it supports five rate-limiting algorithms:
- Token Bucket
- Leaky Bucket
- Fixed Window Counter
- Sliding Window Log
- Sliding Window Counter

It is optimized for high-performance operation under heavy load and is easy to deploy and configure. It can run as a standalone service via Docker and provides a RESTful API to determine whether a request is allowed in real-time.

## Installation
```bash
# Homebrew
homebrew install gate-limiter

# NPM
npm install -g @sapiensxxv/gate-limiter

# with docker compose
git clone https://github.com/your-org/gate-limiter.git
cd gate-limiter/docker
export GATE_LIMITER_TAG=v0.1.0  # or skip this
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

- When using docker compose:
    - The config.yml file is already included in the container.
    - The `GATE_LIMITER_CONFIG` environment variable is defined in docker-compose.yml.
- When using only the Docker image:
    - You must prepare a config.yml file and mount it into the container.
	- The `GATE_LIMITER_CONFIG` environment variable must point to the config path inside the container.

## Setting
An example of the config.yml file is shown below:
```yml
rateLimiter:  
  strategy: sliding_window_counter  
  # token bucket, leaky bucket, fixed window counter, sliding window counter, sliding_window_log  
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
```

- **rateLimiter**: Root of all configuration for rate limiting.
    - **trategy**: Selects which algorithm to use for rate limiting.
        - `token_bucket`: Token Bucket algorithm
        - `leaky_bucket`: Leaky Bucket algorithm
        - `fixed_window_counter`: Fixed Window Counter
        - `sliding_window_log`: Sliding Window Log
        - `sliding_window_counter`: Sliding Window Counter
    - **identity**: Determines how to identify the user.
        - **key**: User identity source
            - `ipv4`: Identify by IPv4 address
        - **header**: Name of the header to extract identity info
    - **apis**: List of API-specific rate limiting rules.
        - **identifier**: Arbitrary unique string to identify the API
        - **path**: Path matching method
            - **expression**: Determines the type of match
                regex: Regular expression
                plain: Literal string
            - **alue**: Actual path string or regex pattern
        - **method**: HTTP method (e.g., `GET`, `POST`)
        - **limit**: Request threshold
        - **windowSeconds**: Time window (in seconds)
        - **refillSeconds**: Token refill interval (for token/leaky buckets)
        - **expireSeconds**: Time before unused state is cleared
    - **target**: The destination domain for forwarded (allowed) requests

## Algorithm
gate-limiter supports the following five rate-limiting algorithms:
- Token Bucket
- Leaky Bucket
- Fixed Window Counter
- Sliding Window Log
- Sliding Window Counter

You can specify the algorithm using the rateLimiter.strategy field in config.yml.

### Token Bucket

This algorithm consumes a token from the bucket for each incoming request.  
If tokens are available, the request is allowed; otherwise, it is rejected.  
Tokens are refilled at a fixed interval.

<p align="center">
	<img width="900" alt="Screenshot 2025-08-01 03:10:10" src="https://github.com/user-attachments/assets/de6bd04f-9148-4e0f-98d2-60eb393fb75d" />
</p>

Two parameters must be configured:
- Bucket size: Controlled by the `rateLimiter.apis.limit` field in `config.yml`
- Token refill interval: Controlled by the `rateLimiter.apis.refillSeconds` field in `config.yml`

### Leaky Bucket
This algorithm enforces a fixed request processing rate over time. It is implemented using Go's channel mechanism.

When a request arrives, the algorithm checks whether the channel (queue) is full:
- If there is space in the channel, the request is added.
- If the channel is full, the request is dropped.

At fixed intervals, a worker removes and processes requests from the channel.

<p align="center">
	<img width="900" alt="Screenshot 2025-08-01 03:13:23" src="https://github.com/user-attachments/assets/62eaa706-97d0-48b1-bfc5-eae9ef80a902" />
</p>

Two parameters must be configured:

- Queue (channel) size: Controlled by the `rateLimiter.apis.limit` field in `config.yml`
- Request processing interval: Controlled by the `rateLimiter.apis.windowSeconds` field in `config.yml`

### Fixed Window Counter

The timeline is divided into fixed-sized units called "windows," and a counter is assigned to each window.

- Each incoming request increments the counter of the current window by 1.
- If the window's counter value is greater than or equal to the threshold, the incoming request is rejected.
- If the window's counter value is less than the threshold, the request is accepted.

The illustration below shows a case where requests are limited to 3 per minute.
<p align="center">
	<img width="900" alt="Screenshot 2025-08-01 04:20:58" src="https://github.com/user-attachments/assets/098a5d02-880d-4b84-b4e7-24c5d34a2f0a" />
</p>

Two parameters need to be configured:
- Window size: Controlled by the `rateLimiter.apis.limit` field in `config.yml`
- Time window duration: Controlled by the `rateLimiter.apis.windowSeconds` field in `config.yml`

### Sliding Window Log

The Sliding Window Logging algorithm is a time-based rate limiting method that stores the exact timestamps of incoming requests within a specified time window to determine whether a new request should be allowed.

- If the log is empty, the request is allowed and its timestamp is recorded.
- If the log is not empty:
  - Remove any timestamps that fall outside the current time window.
  - If the number of timestamps within the window is less than the threshold, the request is allowed.
  - If the number of timestamps is equal to or greater than the threshold, the request is rejected.

<p align="center">
<img width="900" alt="Screenshot 2025-08-01 04:57:03" src="https://github.com/user-attachments/assets/85fc0c83-b11d-43a2-b148-4104781936e1" />
</p>

Two parameters must be configured:

- Window size: Controlled by the `rateLimiter.apis.limit` field in `config.yml`
- Time window duration: Controlled by the `rateLimiter.apis.windowSeconds` field in `config.yml`

### Sliding Window Counter

The Sliding Window Counter algorithm is a hybrid of the Fixed Window and Sliding Window methods.  
It estimates the number of requests in the current window by interpolating between the counts of the current and previous fixed windows, based on how far the current time has progressed into the current window.

<p align="center">
<img width="900" alt="Screenshot 2025-08-01 05:09:27" src="https://github.com/user-attachments/assets/fbe86474-8de0-47ee-bd16-554a7e358d80" />
</p>

Two parameters must be configured:
- Window size: Controlled by the `rateLimiter.apis.limit` field in `config.yml`
- Time window duration: Controlled by the `rateLimiter.apis.windowSeconds` field in `config.yml`

## More Info

If you’re unsure which algorithm to choose, refer to the Rate [Limiter Design Post](https://sapiensxxv.github.io/posts/%EC%B2%98%EB%A6%AC%EC%9C%A8-%EC%A0%9C%ED%95%9C%EA%B8%B0-%EA%B0%9C%EB%B0%9C/).

Author
- Jaehoon So
- Email: jhspacelover@naver.com

## License
gate-limiter is available under the MIT license. See the LICENSE file for more info.
