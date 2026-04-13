package http

import (
	"net/http"
	"order/internal/usecase"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.uc.Create(req.CustomerID, req.ItemName, req.Amount)
	if err != nil {
		if err.Error() == "payment service unavailable" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, OrderResponse{
		ID:         res.ID,
		CustomerID: res.CustomerID,
		ItemName:   res.ItemName,
		Amount:     res.Amount,
		Status:     res.Status,
		CreatedAt:  time.Now(),
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	res, err := h.uc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, OrderResponse{
		ID:         res.ID,
		CustomerID: res.CustomerID,
		ItemName:   res.ItemName,
		Amount:     res.Amount,
		Status:     res.Status,
		CreatedAt:  res.CreatedAt,
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.uc.Cancel(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Cancelled"})
}

func (h *OrderHandler) GetRecent(c *gin.Context) {
	limit := c.Query("limit")
	if limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit parameter is required"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a valid number"})
		return
	}

	orders, err := h.uc.GetRecent(limitInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}
