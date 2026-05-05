package entities

import "time"

type Order struct {
	ID            string
	CustomerID    string
	ItemName      string
	CustomerEmail string
	Amount        int64
	Status        string
	CreatedAt     time.Time
}
