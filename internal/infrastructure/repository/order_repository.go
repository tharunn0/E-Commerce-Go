package repository

import (
	// "context"

	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/tharunn0/E-Commerce-Go/internal/domain"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return OrderRepository{DB: db}
}
