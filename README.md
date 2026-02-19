<h1 align="center">Gate Limiter</h1>

English | [한국어](./README-KR.md)

[![Golang](https://img.shields.io/badge/Go-1.23.3-00ADD8?style=flat&logo=Go)](https://go.dev/doc/)
[![NPM](https://img.shields.io/badge/npm-reference-CB3837?style=flat&logo=npm&logoColor=CB3837&labelColor=747474)](https://www.npmjs.com/package/@sapiensxxv/gate-limiter-cli)
![HomeBrew](https://img.shields.io/badge/Homebrew-reference-FBB040?style=flat&logo=Homebrew&logoColor=FBB040)
[![Docker](https://img.shields.io/badge/Docker-reference-2496ED?style=flat&logo=Docker&logoColor=2496ED)](https://hub.docker.com/repository/docker/sjhn/gate-limiter/general)

---

## Introduction

**gate-limiter** is a configurable rate-limiting middleware designed to prevent API abuse and ensure fair resource usage among users.  
It is written in Go and provides the following five rate-limiting algorithms:

- Token Bucket  
- Leaky Bucket  
- Fixed Window Counter  
- Sliding Window Log  
- Sliding Window Counter  

It is optimized for stability in high-load environments, easy deployment, and flexible configuration.  
You can run it as a standalone service using Docker, and determine request allowance in real time via a RESTful API.

---

## Installation

```bash
# Homebrew
homebrew install gate-limiter

# NPM
npm install -g @sapiensxxv/gate-limiter-cli

# with docker compose
git clone https://github.com/your-org/gate-limiter.git
cd gate-limiter/docker
export GATE_LIMITER_TAG=v0.1.0  # optional
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

### Notes

* **Using NPM**

  * The `config.yml` file must exist in the current directory where the command is executed.
  * Run the rate limiter using the `gate-limiter` command.

* **Using docker compose**

  * The `config.yml` file is included inside the container.
  * The `GATE_LIMITER_CONFIG` environment variable is already defined in `docker-compose.yml`.

* **Using docker image**

  * You must prepare a `config.yml` file and mount it into the container.
  * The `GATE_LIMITER_CONFIG` environment variable must point to the config path inside the container.

---

## Configuration

### Root Configuration (`RootRateLimiterConfig`)

| Key           | Type                                    | Description                      |
| ------------- | --------------------------------------- | -------------------------------- |
| `rateLimiter` | [RateLimiterConfig](#ratelimiterconfig) | Rate limiting configuration root |
| `redis`       | [RedisClientConfig](#redisclientconfig) | Redis client configuration       |

---

### RateLimiterConfig

| Key        | Type                              | Description                                                                                                        |
| ---------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `strategy` | `string`                          | Algorithm (`token_bucket`, `leaky_bucket`, `fixed_window_counter`, `sliding_window_counter`, `sliding_window_log`) |
| `identity` | [ClientIdentity](#clientidentity) | Client identification method                                                                                       |
| `client`   | [ClientLimit](#clientlimit)       | Global per-client rate limit                                                                                       |
| `apis`     | `[Api]`                           | Per-API rate limit rules                                                                                           |
| `target`   | `string`                          | Target domain URL for allowed requests                                                                             |
| `port`     | `int`                             | Server port (default: `8081`)                                                                                      |

---

### ClientIdentity

| Key      | Type     | Description                                          |
| -------- | -------- | ---------------------------------------------------- |
| `key`    | `string` | Identity key (`ipv4` or `cookie`)                    |
| `header` | `string` | HTTP header name for identity value (ipv4 only)      |

---

### ClientLimit

| Key             | Type  | Description            |
| --------------- | ----- | ---------------------- |
| `limit`         | `int` | Max requests           |
| `windowSeconds` | `int` | Time window in seconds |

---

### Api

| Key             | Type                                | Description                    |
| --------------- | ----------------------------------- | ------------------------------ |
| `identifier`    | `string`                            | Unique API identifier          |
| `path`          | [RateLimiterPath](#ratelimiterpath) | API path definition            |
| `method`        | `string`                            | HTTP method                    |
| `limit`         | `int`                               | Request limit                  |
| `windowSeconds` | `int`                               | Window duration                |
| `refillSeconds` | `int`                               | Token refill interval          |
| `expireSeconds` | `int`                               | Storage expiration time        |
| `target`        | `string`                            | Optional per-API target domain |

---

### RateLimiterPath

| Key          | Type     | Description                  |
| ------------ | -------- | ---------------------------- |
| `expression` | `string` | Path type (`regex`, `plain`) |
| `value`      | `string` | Path value                   |

---

### RedisClientConfig

| Key        | Type     | Description    |
| ---------- | -------- | -------------- |
| `host`     | `string` | Redis host     |
| `port`     | `int`    | Redis port     |
| `password` | `string` | Redis password |
| `db`       | `int`    | Redis DB index |

---

## Example Config Scenarios

Each configuration example below represents a real-world usage pattern.

1. Global client limit + API-specific protection  
2. Sensitive API protection (login/payment)  
3. High-throughput public API (burst traffic allowed)  
4. Queue-based traffic shaping for heavy APIs  


### 1. Global client rate limiting + API-specific restriction

**Use case:**
Limit overall user traffic to 50 requests per minute, while restricting the comment creation API to 5 requests per minute to prevent abuse.

```yml
rateLimiter:  
  strategy: sliding_window_counter  
  identity:  
    key: ipv4  
    header: X-Forwarded-For  

  # Global per-client limit
  client:
    limit: 50  
    windowSeconds: 60  

  # API-specific limits
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

  # Forward allowed requests to backend service
  target: https://mywebsitedomain.com
  port: 8081  # default

redis:
  host: localhost
  port: 6379
  password:
  db: 0
```

* Each client (identified by IPv4) can make:
  * **Max 50 requests per minute globally**
  * **Max 5 POST requests per minute** to `/api/item/{id}/comment`
* Allowed requests are forwarded to `https://mywebsitedomain.com`
* Redis is used as shared state storage for rate-limiting counters

---

### 2. Strict API protection (no global limit, API-only limits)

**Use case:**
Do not limit global traffic, but protect only sensitive APIs
(e.g. login, signup, payment).

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

* User is identified by HMAC-signed cookie
* Login API:
  * Max **5 attempts per minute**
* Payment API:
  * Max **3 requests per minute**
* No global client limit
* API abuse prevention focused on sensitive endpoints

---

### 3. High-throughput public API (Token Bucket strategy)

**Use case:**
Allow burst traffic while enforcing an average throughput limit
(e.g. public open APIs, external clients).

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
      limit: 100        # bucket size
      refillSeconds: 10 # refill every 10s
      windowSeconds: 10
      expireSeconds: 600

  target: https://public-api.service.com

redis:
  host: localhost
  port: 6379
  db: 0
```

* Burst traffic allowed (up to 100 requests instantly)
* Tokens refill every 10 seconds
* Smooth rate control for public APIs
* Ideal for CDN-backed APIs or open services

---

### 4. Queue-based traffic shaping (Leaky Bucket)

**Use case:**
Do not drop requests, but enforce a fixed processing rate
(e.g. ML APIs, heavy processing endpoints).

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

**Meaning:**

* Requests are queued
* Processed at a fixed rate
* Overflow requests are dropped
* Traffic smoothing for heavy workloads

## Algorithms

Gate Limiter supports the following algorithms:

* Token Bucket
* Leaky Bucket
* Fixed Window Counter
* Sliding Window Log
* Sliding Window Counter

Configure using:

```yml
rateLimiter.strategy
```

---

### Token Bucket

Consumes tokens per request. Requests pass if tokens exist; otherwise rejected. Tokens refill periodically.

Parameters:

* Bucket size → `rateLimiter.apis.limit`
* Refill interval → `rateLimiter.apis.refillSeconds`

---

### Leaky Bucket

Processes requests at a fixed rate using a queue (Go channel model).
Requests are dropped when the queue is full.

Parameters:

* Queue size → `rateLimiter.apis.limit`
* Processing interval → `rateLimiter.apis.windowSeconds`

---

### Fixed Window Counter

Splits time into fixed windows and counts requests per window.

Rules:

* Increment counter per request
* Reject if counter ≥ limit
* Accept if counter < limit

Parameters:

* Window size → `rateLimiter.apis.limit`
* Window duration → `rateLimiter.apis.windowSeconds`

---

### Sliding Window Log

Stores timestamps of requests and checks counts within a moving time window.

Rules:

* Remove expired timestamps
* Allow if count < limit
* Reject if count ≥ limit

Parameters:

* Window size → `rateLimiter.apis.limit`
* Window duration → `rateLimiter.apis.windowSeconds`

---

### Sliding Window Counter

Hybrid of fixed window and sliding window using weighted approximation between previous and current windows.

Parameters:

* Window size → `rateLimiter.apis.limit`
* Window duration → `rateLimiter.apis.windowSeconds`

---

## Author

* Jaehoon So
* Email: [jhspacelover@naver.com](mailto:jhspacelover@naver.com)

---

## License

`gate-limiter` is available under the **MIT License**.
See the [LICENSE](https://github.com/sapiensXXV/gate-limiter/blob/main/LICENSE) file for more information.
