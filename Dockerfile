FROM golang:1.24-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY config.yaml /app/config.yaml
RUN go build -ldflags="-s -w" -o /app/app cmd/main.go

RUN apk add --no-cache git
RUN git clone https://github.com/grpc-ecosystem/grpc-health-probe /grpc-health-probe

WORKDIR /grpc-health-probe

RUN go build -ldflags="-s -w" -o /usr/local/bin/grpc-health-probe main.go

FROM alpine AS runner

LABEL stage=runner

RUN addgroup -S kvstore \
    && adduser -S kvstore -G kvstore

USER kvstore

WORKDIR /home/kvuser/

COPY --from=builder /app/app app
COPY --from=builder /app/config.yaml config.yaml
COPY --from=builder /usr/local/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

RUN mkdir -p /home/kvstore/data \
    && chown -R kvstore:kvstore /home/kvstore/data

ENV KVSTORE_DATA=/home/kvstore/data

ENTRYPOINT ["./app", "-config", "config.yaml"]