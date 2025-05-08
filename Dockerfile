FROM golang:1.24-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY config.yaml /app/config.yaml
RUN go build -ldflags="-s -w" -o /app/app cmd/main.go

FROM alpine AS runner

ARG ARCH=amd64

ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.38/grpc_health_probe-linux-${ARCH} /usr/local/bin/grpc-health-probe
RUN chmod +x /usr/local/bin/grpc-health-probe

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

USER appuser

WORKDIR /home/appuser/

COPY --from=builder /app/app app
COPY --from=builder /app/config.yaml config.yaml

RUN mkdir -p /home/appuser/data \
    && chown -R appuser:appgroup /home/appuser/data

ENV KVSTORE_DATA=/home/appuser/data

ENTRYPOINT ["./app", "-config", "/home/appuser/config.yaml"]