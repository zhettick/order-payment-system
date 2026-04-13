package grpc

import (
	"payment/internal/domain/entities"

	base "github.com/zhettick/order-payment-gen/payment/v1/base"
)

func toProto(res *entities.Payment) *base.Payment {
	return &base.Payment{
		Id:            res.ID,
		OrderId:       res.OrderID,
		TransactionId: res.TransactionID,
		Amount:        res.Amount,
		Status:        mapStatus(res.Status),
	}
}

func mapStatus(status string) base.PaymentStatus {
	switch status {
	case entities.StatusAuthorized:
		return base.PaymentStatus_PAYMENT_STATUS_AUTHORIZED

	case entities.StatusDeclined:
		return base.PaymentStatus_PAYMENT_STATUS_DECLINED

	default:
		return base.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}
