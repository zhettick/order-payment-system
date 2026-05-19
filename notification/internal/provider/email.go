package provider

import "context"

type EmailMessage struct {
	EventID string
	To      string
	OrderID string
	Amount  int64
	Status  string
}

type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}
