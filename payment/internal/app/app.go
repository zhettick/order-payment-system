package app

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"payment/internal/repository/postgres"
	"payment/internal/transport/grpc"
	"payment/internal/transport/http"
	"payment/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	svc "github.com/zhettick/order-payment-gen/payment/v1/service"
	googlegrpc "google.golang.org/grpc"
)

func Run() {
	dbURL := os.Getenv("PAYMENT_DB_URL")
	grpcPort := os.Getenv("PAYMENT_GRPC_PORT")
	httpPort := os.Getenv("PAYMENT_HTTP_PORT")

	if grpcPort == "" {
		grpcPort = "50051"
	}
	if httpPort == "" {
		httpPort = "8081"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to Payment DB: %v", err)
	}
	defer db.Close()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Payment gRPC failed to listen: %v", err)
		}

		loggingInterceptor := func(ctx context.Context, req interface{}, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler) (interface{}, error) {
			start := time.Now()

			resp, err := handler(ctx, req)

			log.Printf("gRPC Method: %s | Duration: %v | Error: %v", info.FullMethod, time.Since(start), err)
			return resp, err
		}

		s := googlegrpc.NewServer(googlegrpc.UnaryInterceptor(loggingInterceptor))
		paymentGRPCServer := grpc.NewServer(paymentUC)
		svc.RegisterPaymentServiceServer(s, paymentGRPCServer)

		log.Printf("Payment gRPC Server running on port %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Payment gRPC failed to serve: %v", err)
		}
	}()

	r := gin.Default()
	paymentHandler := http.NewPaymentHandler(paymentUC)
	http.SetupRoutes(r, paymentHandler)

	log.Printf("Payment HTTP Server running on port %s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatalf("Payment HTTP server failed: %v", err)
	}
}
