package cli

const defaultConfigTemplate = `# gate-limiter configuration
# https://github.com/sapiensXXV/gate-limiter

rateLimiter:
  # Algorithm: token_bucket | leaky_bucket | fixed_window_counter
  #            | sliding_window_counter | sliding_window_log
  strategy: sliding_window_counter

  identity:
    key: ipv4          # ipv4 | cookie
    header: X-Forwarded-For

  # Global per-client limit
  client:
    limit: 100
    windowSeconds: 60

  # Per-API limits
  apis:
    - identifier: example_api
      path:
        expression: regex      # exact | prefix | regex
        value: ^/api/example$
      method: GET
      limit: 10
      windowSeconds: 60
      refillSeconds: 60        # token_bucket only
      expireSeconds: 3600

  # Upstream target (leave empty for standalone mode)
  target: ""

  # port: 8081       # default 8081
  # adminPort: 8082  # default 8082

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
`
