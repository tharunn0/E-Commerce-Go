package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductFilter_IsFiltersValid(t *testing.T) {
	tests := []struct {
		name   string
		filter *ProductFilter
		want   bool
	}{
		{
			name: "Valid Sort and Order",
			filter: &ProductFilter{
				Sort:  "name",
				Order: "ASC",
			},
			want: true,
		},
		{
			name: "Valid Case Insensitive Sort and Order",
			filter: &ProductFilter{
				Sort:  "NAME",
				Order: "desc",
			},
			want: true,
		},
		{
			name: "Empty Sort and Order",
			filter: &ProductFilter{
				Sort:  "",
				Order: "",
			},
			want: true,
		},
		{
			name: "Price Sort Conversion",
			filter: &ProductFilter{
				Sort:  "price",
				Order: "ASC",
			},
			want: true,
		},
		{
			name: "Invalid Sort Column",
			filter: &ProductFilter{
				Sort:  "invalid",
				Order: "ASC",
			},
			want: false,
		},
		{
			name: "Invalid Order",
			filter: &ProductFilter{
				Sort:  "name",
				Order: "INVALID",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.IsFiltersValid()
			assert.Equal(t, tt.want, got)
			if tt.want && tt.filter.Sort != "" {
				// check normalization
				if tt.filter.Sort == "price" {
					assert.Equal(t, "min_price", tt.filter.Sort)
				}
			}
		})
	}
}
