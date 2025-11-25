package domain

import (
	"fmt"
	"strings"
	"time"
)

type DeliveryType string

const (
	DeliveryTypeNormal  DeliveryType = "NORMAL"
	DeliveryTypeExpress DeliveryType = "EXPRESS"
)

var DeliveryTypeCharges = map[DeliveryType]float64{
	DeliveryTypeNormal:  50,
	DeliveryTypeExpress: 100,
}

var DistrictGroups = map[string][]string{
	"closest": {
		"Ernakulam", "Thrissur", "Kottayam", "Alappuzha",
	},
	"mid": {
		"Pathanamthitta", "Kollam", "Palakkad", "Malappuram", "Idukki",
	},
	"farthest": {
		"Trivandrum", "Kozhikode", "Kannur", "Wayanad", "Kasaragod",
	},
}
var ShippingCharges = map[string]float64{
	"closest":  40,
	"mid":      60,
	"farthest": 80,
}
var DeliveryDays = map[string]int{
	"closest":  2,
	"mid":      3,
	"farthest": 5,
}

func ValidateDistrict(district string) error {
	for _, list := range DistrictGroups {
		for _, item := range list {
			if strings.EqualFold(item, district) {
				return nil
			}
		}
	}
	return fmt.Errorf("unknown district: %s", district)
}

func GetDistrictGroup(district string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(district))

	for group, list := range DistrictGroups {
		for _, item := range list {
			if strings.EqualFold(item, d) {
				return group, nil
			}
		}
	}

	return "", fmt.Errorf("unknown district: %s", district)
}

func GetDeliveryDays(district string) (int, error) {
	group, err := GetDistrictGroup(district)
	if err != nil {
		return 0, err
	}

	return DeliveryDays[group], nil
}

func CalculateShippingCharge(district string) (float64, error) {
	group, err := GetDistrictGroup(district)
	if err != nil {
		return 0, err
	}

	return ShippingCharges[group], nil
}

func CalculateDeliveryDate(district string) (time.Time, error) {
	group, err := GetDistrictGroup(district)
	if err != nil {
		return time.Time{}, err
	}

	days := DeliveryDays[group]
	return time.Now().AddDate(0, 0, days), nil
}
