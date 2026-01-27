package utils

import (
	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

func ValidateUserAddress(address *domain.UserAddress, userId int64) error {
	if address.UserID != userId {
		return apperror.ErrAddressNotFoundForUser
	}
	return nil
}

func ValidateOrderItemsStock(variantMap map[int64]domain.VariantInfo, cart *domain.Cart) []domain.NotEnoughStockError {

	var notEnoughStockErrors []domain.NotEnoughStockError

	for _, cartItem := range cart.Items {
		if variantMap[cartItem.ProductVariantID].Stock < cartItem.Quantity {
			notEnoughStockErrors = append(notEnoughStockErrors, domain.NotEnoughStockError{
				ProductVariantID: cartItem.ProductVariantID,
				SKU:              variantMap[cartItem.ProductVariantID].SKU,
				Quantity:         cartItem.Quantity,
				Stock:            variantMap[cartItem.ProductVariantID].Stock,
			})
		}
	}
	return nil
}

func CalculateOrderTotalAmount(order *domain.OrderResponse) *domain.OrderResponse {
	deductableAmount := 0.0
	for _, v := range order.Items {
		if v.Status == "cancelled" {
			deductableAmount += v.TotalPrice
		}
	}

	order.ShippingCost = domain.DeliveryTypeCharges[order.DeliveryType]
	order.TotalAmount = order.Subtotal + order.TaxAmount + order.ShippingCost
	order.PayableAmount = order.Subtotal + order.TaxAmount + order.ShippingCost - deductableAmount

	return order
}
