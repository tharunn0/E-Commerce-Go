package domain

func ApplyDiscounts(cart *Cart, offers []*Offer) {
	var cartTotal float64

	for _, item := range cart.Items {

		// 1️⃣ Base prices
		unitPrice := item.OriginalPrice
		baseTotal := unitPrice * float64(item.Quantity)

		var bestDiscount float64

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
			}

			if discount > 0 {
				item.AppliedOffer = &AppliedOfferData{
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
		} else {
			item.SalePrice = nil
			item.TotalPrice = baseTotal
		}

		// 6️⃣ Accumulate cart total
		cartTotal += item.TotalPrice
	}

	cart.CartTotalPrice = cartTotal
}

func containsId(arr []int64, val int64) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
