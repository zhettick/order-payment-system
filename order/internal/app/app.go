package app

import (
	"database/sql"
	"log"
	"net"
	"os"

	"order/internal/repository/postgres"
	ordergrpc "order/internal/transport/grpc"
	grpcclient "order/internal/transport/grpc/client"
	"order/internal/transport/http"
	"order/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	svc "github.com/zhettick/order-payment-gen/order/v1/service"
	paymentSvc "github.com/zhettick/order-payment-gen/payment/v1/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func Run() {
	dbURL := os.Getenv("ORDER_DB_URL")
	httpPort := os.Getenv("ORDER_HTTP_PORT")
	grpcPort := os.Getenv("ORDER_GRPC_PORT")
	paymentAddr := os.Getenv("PAYMENT_SERVICE_ADDR")

	if dbURL == "" {
		log.Fatal("ORDER_DB_URL is required")
	}
	if httpPort == "" {
		log.Fatal("ORDER_HTTP_PORT is required")
	}
	if grpcPort == "" {
		log.Fatal("ORDER_GRPC_PORT is required")
	}
	if paymentAddr == "" {
		log.Fatal("PAYMENT_SERVICE_ADDR is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	conn, err := grpc.Dial(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect payment grpc: %v", err)
	}
	defer conn.Close()

	paymentGRPC := paymentSvc.NewPaymentServiceClient(conn)
	paymentClient := grpcclient.NewPaymentGRPCClient(paymentGRPC)

	orderRepo := postgres.NewOrderRepository(db)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("gRPC failed to listen: %v", err)
		}

		s := grpc.NewServer()
		orderGRPCServer := ordergrpc.NewServer(orderUC)
		log.Println("Registering OrderService...")
		svc.RegisterOrderServiceServer(s, orderGRPCServer)
		reflection.Register(s)

		log.Printf("gRPC server listening on port %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC failed to serve: %v", err)
		}
	}()

	r := gin.Default()
	orderHandler := http.NewOrderHandler(orderUC)
	http.SetupRoutes(r, orderHandler)

	log.Printf("HTTP server listening on port %s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
