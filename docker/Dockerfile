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

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# CGO_ENABLED=0 으로 정적 링크, -s -w로 사이즈 줄임
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
      -X gate-limiter/internal/buildinfo.Version=${VERSION} \
      -X gate-limiter/internal/buildinfo.Commit=${COMMIT} \
      -X gate-limiter/internal/buildinfo.BuildDate=${BUILD_DATE}" \
    -o /gl ./cmd/gl

# 런타임 스테이지: 경량 이미지
FROM gcr.io/distroless/static-debian12

COPY --from=builder /gl /usr/local/bin/gl

EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/gl", "run"]
