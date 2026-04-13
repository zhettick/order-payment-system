package http

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, h *PaymentHandler) {
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:order_id", h.GetPayment)
}
