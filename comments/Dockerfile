FROM golang:1.13-alpine AS builder

WORKDIR comments/build
COPY . .

RUN go build -o /build .

FROM alpine as release

RUN apk add --no-cache ca-certificates
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe
WORKDIR /comments
COPY --from=builder /build ./server
EXPOSE 5050
ENTRYPOINT ["/comments/server"]