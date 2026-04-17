package app

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"payment/internal/repository/postgres"
	grpcserver "payment/internal/transport/grpc"
	"payment/internal/usecase"

	_ "github.com/lib/pq"
	svc "github.com/zhettick/order-payment-gen/payment/v1/service"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run() {
	dbURL := os.Getenv("PAYMENT_DB_URL")
	grpcPort := os.Getenv("PAYMENT_GRPC_PORT")

	if dbURL == "" {
		log.Fatal("PAYMENT_DB_URL is required")
	}
	if grpcPort == "" {
		log.Fatal("PAYMENT_GRPC_PORT is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to Payment DB: %v", err)
	}
	defer db.Close()

	paymentRepo := postgres.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Payment gRPC failed to listen: %v", err)
	}

	loggingInterceptor := func(
		ctx context.Context,
		req interface{},
		info *googlegrpc.UnaryServerInfo,
		handler googlegrpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Printf("gRPC Method: %s | Duration: %v | Error: %v", info.FullMethod, time.Since(start), err)
		return resp, err
	}

	s := googlegrpc.NewServer(
		googlegrpc.UnaryInterceptor(loggingInterceptor),
	)

	reflection.Register(s)
	paymentGRPCServer := grpcserver.NewServer(paymentUC)
	svc.RegisterPaymentServiceServer(s, paymentGRPCServer)

	log.Printf("Payment gRPC Server running on port %s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Payment gRPC failed to serve: %v", err)
	}
}
