package payment

import (
	"errors"
	"time"
)

type Wallet struct {
	ID        int64     `json:"id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletTransaction struct {
	ID             int64     `json:"id"`
	WalletID       *int64    `json:"wallet_id,omitempty"`
	Amount         float64   `json:"amount"`
	RelatedOrderID *int64    `json:"related_order_id,omitempty"`
	Type           string    `json:"type"`
	Remark         *string   `json:"remark,omitempty"`
	BalanceBefore  float64   `json:"balance_before"`
	BalanceAfter   float64   `json:"balance_after"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type WalletResponse struct {
	Wallet       Wallet              `json:"wallet"`
	Transactions []WalletTransaction `json:"transactions"`
}

type TransactionFilter struct {
	Type      string    `form:"type"`
	StartDate time.Time `form:"start_date format=2006-01-02"`
	EndDate   time.Time `form:"end_date format=2006-01-02"`
	Limit     int       `form:"limit"`
	Page      int       `form:"page"`
}

func (f *TransactionFilter) Validate() error {

	// if type is empty ,skip, if not empty and not one of the valid ones return error
	if f.Type != "" {
		if f.Type != "purchase" && f.Type != "refund" && f.Type != "topup" {
			return errors.New("invalid transaction type")
		}
	}
	if f.StartDate.IsZero() && f.EndDate.IsZero() {
		startDate := time.Now().AddDate(0, -6, 0)
		endDate := time.Now()

		f.StartDate = startDate
		f.EndDate = endDate
	}

	if f.StartDate.After(f.EndDate) {
		return errors.New("start date cannot be greater than end date")
	}

	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Limit > 50 {
		f.Limit = 50
	}

	if f.Page <= 0 {
		f.Page = 1
	}

	return nil
}
