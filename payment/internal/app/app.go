package app

import (
	"database/sql"
	"log"
	"payment/internal/repository/postgres"
	"payment/internal/transport/http"
	"payment/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Run() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5433/payment_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := postgres.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	h := http.NewPaymentHandler(uc)

	r := gin.Default()
	http.SetupRoutes(r, h)

	log.Println("Payment Service running on :8081")
	r.Run(":8081")
}
