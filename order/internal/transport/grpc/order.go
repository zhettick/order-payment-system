package grpc

import (
	"order/internal/domain/entities"
	"order/internal/usecase"
	"time"

	"github.com/zhettick/order-payment-gen/order/v1/base"
	svc "github.com/zhettick/order-payment-gen/order/v1/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	svc.UnimplementedOrderServiceServer
	uc *usecase.OrderUseCase
}

func NewServer(uc *usecase.OrderUseCase) *Server {
	return &Server{uc: uc}
}

func (s *Server) SubscribeToOrderUpdates(
	req *svc.SubscribeRequest,
	stream svc.OrderService_SubscribeToOrderUpdatesServer,
) error {
	orderID := req.GetId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order ID is required")
	}

	order, err := s.uc.GetByID(orderID)
	if err != nil {
		return status.Errorf(codes.NotFound, "order not found")
	}

	lastStatus := order.Status

	if err := stream.Send(&base.Order{
		Id:         order.ID,
		CustomerId: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     mapStatusToProto(order.Status),
		CreatedAt:  timestamppb.New(order.CreatedAt),
	}); err != nil {
		return err
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil

		case <-ticker.C:
			current, err := s.uc.GetByID(orderID)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to read order status")
			}

			if current.Status != lastStatus {
				lastStatus = current.Status

				if err := stream.Send(&base.Order{
					Id:         current.ID,
					CustomerId: current.CustomerID,
					ItemName:   current.ItemName,
					Amount:     current.Amount,
					Status:     mapStatusToProto(current.Status),
					CreatedAt:  timestamppb.New(current.CreatedAt),
				}); err != nil {
					return err
				}

				if current.Status == entities.StatusPaid ||
					current.Status == entities.StatusCancelled ||
					current.Status == entities.StatusFailed {
					return nil
				}
			}
		}
	}
}

func mapStatusToProto(s string) base.OrderStatus {
	switch s {
	case entities.StatusPending:
		return base.OrderStatus_ORDER_STATUS_PENDING
	case entities.StatusPaid:
		return base.OrderStatus_ORDER_STATUS_PAID
	case entities.StatusCancelled:
		return base.OrderStatus_ORDER_STATUS_CANCELED
	case entities.StatusFailed:
		return base.OrderStatus_ORDER_STATUS_FAILED
	default:
		return base.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
