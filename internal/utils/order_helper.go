package utils

import (
	"errors"
	"fmt"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
)

var notEnoughStockErrors []domain.NotEnoughStockError

func ValidateUserAddress(address *domain.UserAddress, userId int64) error {
	if address.UserID != userId {
		return apperror.ErrAddressNotFoundForUser
	}
	return nil
}

func ValidateOrderItemsStock(variantMap map[int64]domain.VariantInfo, cart *domain.Cart) []domain.NotEnoughStockError {

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

func ValidateCart(cart *domain.Cart, maxOrderAmount int64) error {

	if cart.Items == nil {
		return errors.New("Cart is empty.")
	}

	if cart.CartTotalPrice > float64(maxOrderAmount) {
		return fmt.Errorf("Order amount exceeded. Should be less than %d.", maxOrderAmount)
	}
	return nil
}

func CreateOrderItems(cart *domain.Cart, cartVariantInfo map[int64]domain.VariantInfo) []domain.OrderItem {
	var items []domain.OrderItem

	for _, cartItem := range cart.Items {
		var item domain.OrderItem
		item.ProductVariantID = cartItem.ProductVariantID
		item.ProductName = cartVariantInfo[cartItem.ProductVariantID].ProductName
		item.SKU = cartVariantInfo[cartItem.ProductVariantID].SKU
		item.Quantity = cartItem.Quantity
		item.TotalPrice = cartItem.TotalPrice
		if cartItem.SalePrice != nil {
			item.UnitPrice = *cartItem.SalePrice
		} else {
			item.UnitPrice = cartItem.OriginalPrice
		}

		item.OfferData = cartItem.AppliedOffer

		items = append(items, item)
	}

	return items
}
