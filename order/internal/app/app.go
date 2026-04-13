package app

import (
	"database/sql"
	"log"
	"order/internal/repository/postgres"
	"order/internal/transport/http"
	"order/internal/transport/http/client"
	"order/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Run() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5431/order_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := postgres.NewOrderRepository(db)
	pClient := client.NewPaymentClient("http://localhost:8081")
	uc := usecase.NewOrderUseCase(repo, pClient)
	h := http.NewOrderHandler(uc)

	r := gin.Default()
	http.SetupRoutes(r, h)

	r.Run(":8080")
}
