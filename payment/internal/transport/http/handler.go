package http

import (
	"net/http"
	"payment/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.uc.Process(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PaymentResponse{
		ID:            res.ID,
		OrderID:       res.OrderID,
		TransactionID: res.TransactionID,
		Amount:        res.Amount,
		Status:        res.Status,
	})
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	orderID := c.Param("order_id")
	res, err := h.uc.GetByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, PaymentResponse{
		ID:            res.ID,
		OrderID:       res.OrderID,
		TransactionID: res.TransactionID,
		Amount:        res.Amount,
		Status:        res.Status,
	})
}
