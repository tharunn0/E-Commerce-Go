package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTopSellingRequest_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		req     *TopSellingRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Type product",
			req: &TopSellingRequest{
				Type: "product",
				From: now.Add(-10 * 24 * time.Hour),
				To:   now,
			},
			wantErr: false,
		},
		{
			name: "Valid Type category",
			req: &TopSellingRequest{
				Type: "category",
			},
			wantErr: false,
		},
		{
			name: "Valid Type brand",
			req: &TopSellingRequest{
				Type: "brand",
			},
			wantErr: false,
		},
		{
			name: "Invalid Type",
			req: &TopSellingRequest{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name: "From Date in Future",
			req: &TopSellingRequest{
				Type: "product",
				From: future,
			},
			wantErr: true,
			errMsg:  "from date cannot be in the future",
		},
		{
			name: "From Date After To Date",
			req: &TopSellingRequest{
				Type: "product",
				From: now.Add(-5 * 24 * time.Hour),
				To:   now.Add(-10 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "from date cannot be after to date",
		},
		{
			name: "To Date in Future",
			req: &TopSellingRequest{
				Type: "product",
				From: now.Add(-5 * 24 * time.Hour),
				To:   future,
			},
			wantErr: true,
			errMsg:  "to date cannot be in the future",
		},
		{
			name: "Invalid Limit Too Low",
			req: &TopSellingRequest{
				Type:  "product",
				Limit: -1,
			},
			wantErr: true,
			errMsg:  "limit must be between 1 and 10",
		},
		{
			name: "Invalid Limit Too High",
			req: &TopSellingRequest{
				Type:  "product",
				Limit: 20,
			},
			wantErr: true,
			errMsg:  "limit must be between 1 and 10",
		},
		{
			name: "Default Limit Set",
			req: &TopSellingRequest{
				Type:  "product",
				Limit: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(now)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tt.req.Limit == 0 && tt.name == "Default Limit Set" {
					assert.Equal(t, 10, tt.req.Limit)
				}
			}
		})
	}
}

func TestSalesReportRequest_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		req     *SalesReportRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Request",
			req: &SalesReportRequest{
				From: now.Add(-10 * 24 * time.Hour),
				To:   now,
			},
			wantErr: false,
		},
		{
			name: "Default Dates",
			req: &SalesReportRequest{},
			wantErr: false,
		},
		{
			name: "From Date in Future",
			req: &SalesReportRequest{
				From: future,
			},
			wantErr: true,
			errMsg:  "from date cannot be in the future",
		},
		{
			name: "From Date After To Date",
			req: &SalesReportRequest{
				From: now.Add(-5 * 24 * time.Hour),
				To:   now.Add(-10 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "from date cannot be after to date",
		},
		{
			name: "To Date in Future",
			req: &SalesReportRequest{
				From: now.Add(-10 * 24 * time.Hour),
				To:   future,
			},
			wantErr: true,
			errMsg:  "to date cannot be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(now)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRevenueAnalyticsRequest_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		req     *RevenueAnalyticsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Request",
			req: &RevenueAnalyticsRequest{
				Interval: "monthly",
			},
			wantErr: false,
		},
		{
			name: "Default Interval",
			req: &RevenueAnalyticsRequest{},
			wantErr: false,
		},
		{
			name: "Invalid Interval",
			req: &RevenueAnalyticsRequest{
				Interval: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid interval",
		},
		{
			name: "From Date in Future",
			req: &RevenueAnalyticsRequest{
				From: future,
			},
			wantErr: true,
			errMsg:  "from date cannot be in the future",
		},
		{
			name: "From Date After To Date",
			req: &RevenueAnalyticsRequest{
				From: now.Add(-5 * 24 * time.Hour),
				To:   now.Add(-10 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "from date cannot be after to date",
		},
		{
			name: "To Date in Future",
			req: &RevenueAnalyticsRequest{
				From: now.Add(-10 * 24 * time.Hour),
				To:   future,
			},
			wantErr: true,
			errMsg:  "to date cannot be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(now)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tt.name == "Default Interval" {
					assert.Equal(t, "daily", tt.req.Interval)
				}
			}
		})
	}
}
