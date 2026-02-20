# 빌드 스테이지: 정적 바이너리 생성
FROM golang:1.24.5-bookworm AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /src

# 모듈 캐시 다운받기
COPY go.mod go.sum ./
RUN go mod download

# 소스 복사 및 빌드
COPY . .
# CGO_ENABLED=0 으로 정적 링크, -s -w로 사이즈 줄임
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X gate-limiter/internal/cli.Version=${VERSION} -X gate-limiter/internal/cli.Commit=${COMMIT} -X gate-limiter/internal/cli.BuildDate=${BUILD_DATE}" \
    -o /gl ./cmd/gl

# 런타임 스테이지: 경량 이미지
FROM gcr.io/distroless/static-debian12

COPY --from=builder /gl /usr/local/bin/gl

# 비루트 실행을 원하면 사용자 추가하는 단계 필요 (distroless는 이미 제한적)
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/gl", "run"]
