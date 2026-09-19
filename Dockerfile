FROM golang:1.26-alpine AS builder

WORKDIR /build

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/rpc-agent .

FROM alpine:3.20

ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app
COPY --from=builder /app/rpc-agent /app/rpc-agent
COPY etc /app/etc

EXPOSE 8087

ENTRYPOINT ["/app/rpc-agent"]
CMD ["-f", "etc/agent.yaml"]
