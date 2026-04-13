package grpc

import (
	"context"
	"payment/internal/usecase"

	svc "github.com/zhettick/order-payment-gen/payment/v1/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	svc.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewServer(uc *usecase.PaymentUseCase) *Server {
	return &Server{
		uc: uc,
	}
}

func (s *Server) Create(ctx context.Context, req *svc.CreateRequest) (*svc.CreateResponse, error) {
	res, err := s.uc.Process(req.OrderId, req.Amount)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return &svc.CreateResponse{Payment: toProto(res)}, nil
}

func (s *Server) GetByID(ctx context.Context, req *svc.GetByIDRequest) (*svc.GetByIDResponse, error) {
	res, err := s.uc.GetByID(req.OrderId)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment with orderID %s not found", req.OrderId)
	}

	return &svc.GetByIDResponse{Payment: toProto(res)}, nil
}
