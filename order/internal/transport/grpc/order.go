package grpc

import (
	"order/internal/domain/entities"
	"order/internal/usecase"

	"github.com/zhettick/order-payment-gen/order/v1/base"
	svc "github.com/zhettick/order-payment-gen/order/v1/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// initial state
	if err := stream.Send(&base.Order{
		Id:     order.ID,
		Status: mapStatusToProto(order.Status),
	}); err != nil {
		return err
	}

	ch := s.uc.Subscribe(orderID)
	defer s.uc.Unsubscribe(orderID, ch)

	for {
		select {

		case <-stream.Context().Done():
			return nil

		case newStatus := <-ch:
			if err := stream.Send(&base.Order{
				Id:     orderID,
				Status: mapStatusToProto(newStatus),
			}); err != nil {
				return err
			}

			if newStatus == entities.StatusPaid ||
				newStatus == entities.StatusCancelled ||
				newStatus == entities.StatusFailed {
				return nil
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
