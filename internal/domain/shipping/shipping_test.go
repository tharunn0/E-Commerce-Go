package shipping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDistrict(t *testing.T) {
	tests := []struct {
		district string
		wantErr  bool
	}{
		{"Ernakulam", false},
		{"THRISSUR", false},
		{"Unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.district, func(t *testing.T) {
			err := ValidateDistrict(tt.district)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDistrictGroup(t *testing.T) {
	tests := []struct {
		district string
		expected string
		wantErr  bool
	}{
		{"Ernakulam", "closest", false},
		{"Idukki", "mid", false},
		{"Kasaragod", "farthest", false},
		{"Unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.district, func(t *testing.T) {
			got, err := GetDistrictGroup(tt.district)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestGetDeliveryDays(t *testing.T) {
	tests := []struct {
		district string
		expected int
		wantErr  bool
	}{
		{"Ernakulam", 2, false},
		{"Idukki", 3, false},
		{"Kasaragod", 5, false},
		{"Unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.district, func(t *testing.T) {
			got, err := GetDeliveryDays(tt.district)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestCalculateShippingCharge(t *testing.T) {
	tests := []struct {
		district string
		expected float64
		wantErr  bool
	}{
		{"Ernakulam", 40.0, false},
		{"Idukki", 60.0, false},
		{"Kasaragod", 80.0, false},
		{"Unknown", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.district, func(t *testing.T) {
			got, err := CalculateShippingCharge(tt.district)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestGenerateTrackingID(t *testing.T) {
	assert.Contains(t, GenerateTrackingID(CarrierDHL), "DHL-")
	assert.Contains(t, GenerateTrackingID(CarrierFedEx), "FED-")
	assert.Contains(t, GenerateTrackingID(CarrierBlueDart), "BLU-")
}

func TestCalculateDeliveryDate(t *testing.T) {
	tests := []struct {
		district string
		wantErr  bool
	}{
		{"Ernakulam", false},
		{"Unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.district, func(t *testing.T) {
			got, err := CalculateDeliveryDate(tt.district)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, got.IsZero())
			}
		})
	}
}

func TestSelectRandomCarrier(t *testing.T) {
	carrier := SelectRandomCarrier()
	assert.Contains(t, []Carrier{CarrierBlueDart, CarrierFedEx, CarrierDHL}, carrier)
}

func TestShipmentStatusFlow(t *testing.T) {
    assert.NotNil(t, ShipmentStatusFlow)
    assert.Contains(t, ShipmentStatusFlow[ShipmentStatusPending], ShipmentStatusShipped)
}
