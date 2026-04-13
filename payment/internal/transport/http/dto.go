package http

type CreatePaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type PaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}
