FROM golang:1.26-alpine

WORKDIR /app/payment

COPY payment/go.mod payment/go.sum ./
RUN go mod download

COPY payment/ ./

RUN go build -o /main ./cmd/payment/main.go

EXPOSE 50051

CMD ["/main"]