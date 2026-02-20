package cli

var defaultConfigTemplate = `# gate-limiter configuration

rateLimiter:
  strategy: "fixed_window_counter"
  identity:
    key: "ipv4"
    header: "X-Forwarded-For"
  client:
    limit: 100
    windowSeconds: 60
  target: "http://localhost:3000"
  port: 8081
  adminPort: 8082
  apis:
    - identifier: "example-api"
      path:
        expression: "plain"
        value: "/api/example"
      method: "GET"
      limit: 10
      windowSeconds: 60
      refillSeconds: 1
      expireSeconds: 60

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# Logging configuration (optional)
# logging:
#   level: "info"          # debug | info | warn | error
#   format: "text"         # text | json
#   output: "stdout"       # stdout | stderr | file
#   file:
#     directory: "./logs"  # log directory
#     maxSizeMB: 100       # max file size in MB
#     maxAgeDays: 30       # max retention days
#     compress: false       # gzip compression for rotated files
`
