FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go env -w GOPROXY=https://proxy.golang.org,direct && go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/upcycle-hub ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /out/upcycle-hub ./upcycle-hub
COPY config ./config
COPY frontend ./frontend
RUN mkdir -p /data/uploads /data/db
VOLUME ["/data"]
EXPOSE 8080
ENV UPCYCLE_DB_DSN=/data/db/upcycle.db
ENV UPCYCLE_UPLOAD_DIR=/data/uploads
ENV UPCYCLE_SERVER_MODE=release
ENTRYPOINT ["./upcycle-hub", "config/config.yaml"]
