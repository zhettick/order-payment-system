FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY notification/go.mod notification/go.sum ./
RUN go mod download

COPY notification/ .

RUN go build -o notification-service ./cmd/notification

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/notification-service .

CMD ["./notification-service"]