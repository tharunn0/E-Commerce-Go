package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateReviewRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateReviewRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Request",
			req: &CreateReviewRequest{
				UserID:      1,
				ProductID:   1,
				Rating:      4.5,
				Title:       "Great Product",
				Description: "I really loved this product, it worked as expected.",
			},
			wantErr: false,
		},
		{
			name: "Invalid Rating - Too Low",
			req: &CreateReviewRequest{
				Rating:      0.5,
				Title:       "Great",
				Description: "Great product...",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Invalid rating",
		},
		{
			name: "Invalid Rating - Too High",
			req: &CreateReviewRequest{
				Rating:      5.5,
				Title:       "Great",
				Description: "Great product...",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Invalid rating",
		},
		{
			name: "Invalid Title - Too Short",
			req: &CreateReviewRequest{
				Rating:      4.0,
				Title:       "Good",
				Description: "Great product...",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Title must be between 5 and 50 characters",
		},
		{
			name: "Invalid Title - Too Long",
			req: &CreateReviewRequest{
				Rating:      4.0,
				Title:       "This is a very long title that exceeds the 50 characters limit",
				Description: "Great product...",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Title must be between 5 and 50 characters",
		},
		{
			name: "Invalid Description - Too Short",
			req: &CreateReviewRequest{
				Rating:      4.0,
				Title:       "Great",
				Description: "Bad",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Description must be between 5 and 200 characters",
		},
		{
			name: "Invalid Description - Too Long",
			req: &CreateReviewRequest{
				Rating:      4.0,
				Title:       "Great",
				Description: "This description is definitely way too long and is going to exceed the maximum allowed length of two hundred characters which is set as a limit in the validation function for the CreateReviewRequest struct.",
				ProductID:   1,
			},
			wantErr: true,
			errMsg:  "Description must be between 5 and 200 characters",
		},
		{
			name: "Invalid Product ID",
			req: &CreateReviewRequest{
				Rating:      4.0,
				Title:       "Great",
				Description: "Great product...",
				ProductID:   0,
			},
			wantErr: true,
			errMsg:  "Invalid product ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReviewFilter_Validate(t *testing.T) {
	tests := []struct {
		name       string
		filter     *ReviewFilter
		wantErr    bool
		errMsg     string
		wantFilter *ReviewFilter
	}{
		{
			name: "Valid Filter Rating Asc",
			filter: &ReviewFilter{
				SortBy:  "rating",
				OrderBy: "asc",
				Page:    2,
				Limit:   20,
			},
			wantErr: false,
			wantFilter: &ReviewFilter{
				SortBy:  "created_at",
				OrderBy: "desc",
				Page:    2,
				Limit:   20,
			},
		},
		{
			name: "Valid Filter CreatedAt Desc",
			filter: &ReviewFilter{
				SortBy:  "created_at",
				OrderBy: "desc",
				Page:    2,
				Limit:   20,
			},
			wantErr: false,
			wantFilter: &ReviewFilter{
				SortBy:  "created_at",
				OrderBy: "desc",
				Page:    2,
				Limit:   20,
			},
		},
		{
			name: "Invalid Sort By",
			filter: &ReviewFilter{
				SortBy: "invalid",
			},
			wantErr: true,
			errMsg:  "Invalid sort by",
		},
		{
			name: "Invalid Order By",
			filter: &ReviewFilter{
				SortBy:  "rating",
				OrderBy: "invalid",
			},
			wantErr: true,
			errMsg:  "Invalid order by",
		},
		{
			name: "Default Values",
			filter: &ReviewFilter{
				Page:  0,
				Limit: 0,
			},
			wantErr: false,
			wantFilter: &ReviewFilter{
				SortBy:  "",
				OrderBy: "",
				Page:    1,
				Limit:   10,
			},
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
				assert.Equal(t, tt.wantFilter.SortBy, tt.filter.SortBy)
				assert.Equal(t, tt.wantFilter.OrderBy, tt.filter.OrderBy)
				assert.Equal(t, tt.wantFilter.Page, tt.filter.Page)
				assert.Equal(t, tt.wantFilter.Limit, tt.filter.Limit)
			}
		})
	}
}
