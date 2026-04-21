package postgres

import (
	"database/sql"
	"payment/internal/domain/entities"
	"strconv"

	_ "github.com/lib/pq"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(p *entities.Payment) error {
	query := `INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status)
	return err
}

func (r *PaymentRepository) GetByID(orderID string) (*entities.Payment, error) {
	p := &entities.Payment{}
	query := `SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id = $1`
	err := r.db.QueryRow(query, orderID).Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	return p, err
}

func (r *PaymentRepository) FindByAmountRange(min, max int64) ([]entities.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status FROM payments WHERE 1=1`
	var args []interface{}

	if min > 0 {
		args = append(args, min)
		query += " AND amount >= $" + strconv.Itoa(len(args))
	}

	if max > 0 {
		args = append(args, max)
		query += " AND amount <= $" + strconv.Itoa(len(args))
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []entities.Payment
	for rows.Next() {
		var p entities.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, nil
}
