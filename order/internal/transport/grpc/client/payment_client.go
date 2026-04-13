package client

import (
	"context"
	"log"
	"order/internal/domain/entities"
	"time"

	"github.com/zhettick/order-payment-gen/payment/v1/base"
	svc "github.com/zhettick/order-payment-gen/payment/v1/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCClient struct {
	client svc.PaymentServiceClient
}

func NewPaymentGRPCClient(client svc.PaymentServiceClient) *PaymentGRPCClient {
	return &PaymentGRPCClient{
		client: client,
	}
}

func (c *PaymentGRPCClient) Authorize(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &svc.CreateRequest{
		OrderId: orderID,
		Amount:  amount,
	}

	resp, err := c.client.Create(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("Payment service returned gRPC error: %v - %v", st.Code(), st.Message())
			return entities.StatusFailed, status.Error(st.Code(), "payment service call failed")
		}
	}

	if resp.Payment.Status == base.PaymentStatus_PAYMENT_STATUS_AUTHORIZED {
		return entities.StatusPaid, nil
	}

	return entities.StatusFailed, status.Error(codes.Internal, "unexpected communication error")
}
