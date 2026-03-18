package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateUserProfileRequest_Validate(t *testing.T) {
	tests := []struct {
		name string
		req  *UpdateUserProfileRequest
	}{
		{
			name: "Normalize empty strings to nil",
			req: &UpdateUserProfileRequest{
				FirstName:      func() *string { s := ""; return &s }(),
				LastName:       func() *string { s := ""; return &s }(),
				Phone:          func() *string { s := ""; return &s }(),
				ProfilePicture: func() *string { s := ""; return &s }(),
			},
		},
		{
			name: "Keep non-empty strings",
			req: &UpdateUserProfileRequest{
				FirstName: func() *string { s := "John"; return &s }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.req.Validate()
			if tt.name == "Normalize empty strings to nil" {
				assert.Nil(t, tt.req.FirstName)
				assert.Nil(t, tt.req.LastName)
				assert.Nil(t, tt.req.Phone)
				assert.Nil(t, tt.req.ProfilePicture)
			} else {
				assert.NotNil(t, tt.req.FirstName)
				assert.Equal(t, "John", *tt.req.FirstName)
			}
		})
	}
}
