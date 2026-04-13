package http

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine, h *OrderHandler) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
	r.GET("/orders/recent", h.GetRecent)
}
