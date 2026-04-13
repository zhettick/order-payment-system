FROM golang:1.26-alpine

# Set the working directory to the payment folder
WORKDIR /app/payment

# Copy the payment-specific go files first
COPY payment/go.mod payment/go.sum ./
RUN go mod download

# Copy the rest of the payment code
COPY payment/ ./

# Build from the current directory (which is /app/payment)
# Since the go.mod here says "module payment",
# the import "payment/internal/app" will now work!
RUN go build -o /main ./cmd/payment/main.go

EXPOSE 50051

CMD ["/main"]