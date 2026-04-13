FROM golang:1.26-alpine

WORKDIR /app/order

# Копируем go.mod именно из папки order
COPY order/go.mod order/go.sum ./
RUN go mod download

COPY order/ ./

# Собираем main из папки order
RUN go build -o /main ./cmd/order/main.go

EXPOSE 8080

CMD ["/main"]