package utils

import (
	"errors"
	"fmt"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/order"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/user"
)

var notEnoughStockErrors []order.NotEnoughStockError

func ValidateUserAddress(address *user.UserAddress, userId int64) error {
	if address.UserID != userId {
		return apperror.ErrAddressNotFoundForUser
	}
	return nil
}

func ValidateOrderItemsStock(variantMap map[int64]cart.VariantInfo, cart *cart.Cart) []order.NotEnoughStockError {

	for _, cartItem := range cart.Items {
		if variantMap[cartItem.ProductVariantID].Stock < cartItem.Quantity {
			notEnoughStockErrors = append(notEnoughStockErrors, order.NotEnoughStockError{
				ProductVariantID: cartItem.ProductVariantID,
				SKU:              variantMap[cartItem.ProductVariantID].SKU,
				Quantity:         cartItem.Quantity,
				Stock:            variantMap[cartItem.ProductVariantID].Stock,
			})
		}
	}
	return nil
}

func CalculateOrderTotalAmount(ord *order.OrderResponse) *order.OrderResponse {
	deductableAmount := 0.0
	for _, v := range ord.Items {
		if v.Status == "cancelled" {
			deductableAmount += v.TotalPrice
		}
	}

	ord.ShippingCost = order.DeliveryTypeCharges[ord.DeliveryType]
	ord.TotalAmount = ord.Subtotal + ord.TaxAmount + ord.ShippingCost
	ord.PayableAmount = ord.Subtotal + ord.TaxAmount + ord.ShippingCost - deductableAmount

	return ord
}

func ValidateCart(c *cart.Cart, maxOrderAmount int64) error {

	if c.Items == nil {
		return errors.New("Cart is empty.")
	}

	if c.CartTotalPrice > float64(maxOrderAmount) {
		return fmt.Errorf("Order amount exceeded. Should be less than %d.", maxOrderAmount)
	}
	return nil
}

func CreateOrderItems(c *cart.Cart, cartVariantInfo map[int64]cart.VariantInfo) []order.OrderItem {
	var items []order.OrderItem

	for _, cartItem := range c.Items {
		var item order.OrderItem
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
