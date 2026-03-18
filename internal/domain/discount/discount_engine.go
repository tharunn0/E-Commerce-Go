package discount

import (
	"fmt"

	"github.com/tharunn0/E-Commerce-Go/internal/domain/cart"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/product"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/promotion"
)

func ApplyDiscounts(cart *cart.Cart, offers []*promotion.Offer) {
	var cartTotal float64

	for _, item := range cart.Items {

		// 1️⃣ Base prices
		unitPrice := item.OriginalPrice
		baseTotal := unitPrice * float64(item.Quantity)

		var bestDiscount float64
		var bestOffer *promotion.AppliedOfferData

		// 2️⃣ Evaluate offers
		for _, offer := range offers {

			applies := false

			// product match
			if containsId(offer.ProductIDs, item.ProductID) {
				applies = true
			}

			// category match
			if containsId(offer.CategoryIDs, item.CategoryID) {
				applies = true
			}

			if !applies {
				continue
			}

			// 3️⃣ Calculate discount
			var discount float64
			switch offer.DiscountType {
			case "percentage":
				discount = baseTotal * (offer.DiscountValue / 100)
			case "flat":
				discount = offer.DiscountValue
			default:
				continue
			}

			if discount > bestDiscount {
				bestDiscount = discount
				bestOffer = &promotion.AppliedOfferData{
					OfferID:        offer.ID,
					OfferName:      offer.Name,
					DiscountType:   offer.DiscountType,
					DiscountValue:  offer.DiscountValue,
					DiscountAmount: discount,
				}
			}
		}

		// 4️⃣ Cap discount
		if bestDiscount > baseTotal {
			bestDiscount = baseTotal
		}

		// 5️⃣ Apply result
		if bestDiscount > 0 {
			discountedTotal := baseTotal - bestDiscount
			discountedUnit := discountedTotal / float64(item.Quantity)

			item.SalePrice = &discountedUnit
			item.TotalPrice = discountedTotal
			item.AppliedOffer = bestOffer
			// update bestOffer discount amount if it was capped
			if bestDiscount < bestOffer.DiscountAmount {
				item.AppliedOffer.DiscountAmount = bestDiscount
			}
		} else {
			item.SalePrice = nil
			item.TotalPrice = baseTotal
			item.AppliedOffer = nil
		}

		// 6️⃣ Accumulate cart total
		cartTotal += item.TotalPrice
	}

	cart.CartTotalPrice = cartTotal
}

func ApplyDiscountsToVariants(variants []*product.ProductVariantResponse, offers []*promotion.Offer) {

	fmt.Println("apply discounts to variants called")

	for _, variant := range variants {

		if variant == nil || variant.BaseProduct == nil {
			continue
		}

		unitPrice := variant.OriginalPrice
		baseTotal := unitPrice // quantity = 1

		var bestDiscount float64
		var bestOffer *promotion.AppliedOfferData

		// evaluate offers
		for _, offer := range offers {

			applies := false

			// product match
			if containsId(offer.ProductIDs, variant.BaseProduct.ID) {
				applies = true
			}

			// category match
			if containsId(offer.CategoryIDs, variant.BaseProduct.CategoryID) {
				applies = true
			}

			if !applies {
				continue
			}

			// calculate discount
			var discount float64
			switch offer.DiscountType {
			case "percentage":
				discount = baseTotal * (offer.DiscountValue / 100)
			case "flat":
				discount = offer.DiscountValue
			default:
				continue
			}

			// keep best offer
			if discount > bestDiscount {
				bestDiscount = discount
				bestOffer = &promotion.AppliedOfferData{
					OfferID:        offer.ID,
					OfferName:      offer.Name,
					DiscountType:   offer.DiscountType,
					DiscountValue:  offer.DiscountValue,
					DiscountAmount: discount,
				}
			}
		}

		// cap discount
		if bestDiscount > baseTotal {
			bestDiscount = baseTotal
		}

		// apply result
		if bestDiscount > 0 {
			discountedPrice := baseTotal - bestDiscount
			variant.SalePrice = &discountedPrice
			variant.AppliedOffer = &promotion.AppliedOfferData{
				OfferID:        bestOffer.OfferID,
				OfferName:      bestOffer.OfferName,
				DiscountType:   bestOffer.DiscountType,
				DiscountValue:  bestOffer.DiscountValue,
				DiscountAmount: bestDiscount,
			}
		} else {
			variant.SalePrice = nil
			variant.AppliedOffer = nil
		}
	}
}

func containsId(arr []int64, val int64) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
