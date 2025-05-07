FROM golang:1.24-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux

WORKDIR /build

COPY go.mod ./src/go.sum ./
RUN go mod download

COPY ./src .
RUN go build -ldflags="-s -w" -o /app/app cmd/main.go

FROM alpine AS runner

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

USER appuser

WORKDIR /home/appuser/app

COPY --from=builder /app/app app

ENTRYPOINT ["./app", "--config-path=./config.yaml"]