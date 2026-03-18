package payment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionFilter_Validate(t *testing.T) {
	now := time.Now()
	past := now.AddDate(0, 0, -1)
	future := now.AddDate(0, 0, 1)

	tests := []struct {
		name    string
		filter  *TransactionFilter
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Filter with Type purchase",
			filter: &TransactionFilter{
				Type: "purchase",
			},
			wantErr: false,
		},
		{
			name: "Valid Filter with Type refund",
			filter: &TransactionFilter{
				Type: "refund",
			},
			wantErr: false,
		},
		{
			name: "Valid Filter with Type topup",
			filter: &TransactionFilter{
				Type: "topup",
			},
			wantErr: false,
		},
		{
			name: "Invalid Transaction Type",
			filter: &TransactionFilter{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid transaction type",
		},
		{
			name: "Default Dates if Zero",
			filter: &TransactionFilter{
				StartDate: time.Time{},
				EndDate:   time.Time{},
			},
			wantErr: false,
		},
		{
			name: "Start Date After End Date",
			filter: &TransactionFilter{
				StartDate: future,
				EndDate:   past,
			},
			wantErr: true,
			errMsg:  "start date cannot be greater than end date",
		},
		{
			name: "Limit Normalization",
			filter: &TransactionFilter{
				Limit: 0,
			},
			wantErr: false,
		},
		{
			name: "Limit Cap",
			filter: &TransactionFilter{
				Limit: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tt.name == "Default Dates if Zero" {
					assert.False(t, tt.filter.StartDate.IsZero())
					assert.False(t, tt.filter.EndDate.IsZero())
				}
				if tt.name == "Limit Normalization" {
					assert.Equal(t, 10, tt.filter.Limit)
				}
				if tt.name == "Limit Cap" {
					assert.Equal(t, 50, tt.filter.Limit)
				}
			}
		})
	}
}
