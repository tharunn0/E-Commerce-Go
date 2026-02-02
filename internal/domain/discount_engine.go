package domain

import "time"

func ApplyBestOffers(cart *Cart, offers []Offer) *Cart {
	now := time.Now()

	// Build fast lookup maps once (O(offers))
	productMap := buildProductOfferMap(offers)
	categoryMap := buildCategoryOfferMap(offers)

	var cartTotal float64

	// Process each cart item (O(items))
	for _, item := range cart.Items {

		basePrice := item.OriginalPrice
		bestDiscount := 0.0

		// Collect all applicable offers for this item
		candidates := gatherOffers(item, productMap, categoryMap)

		// Find best discount among candidates
		for _, offer := range candidates {
			if !isOfferValid(offer, now) {
				continue
			}

			discount := calculateDiscountAmount(basePrice, offer)

			if discount > bestDiscount {
				bestDiscount = discount
			}
		}

		// Apply final price
		finalPrice := basePrice - bestDiscount
		if finalPrice < 0 {
			finalPrice = 0
		}

		item.SalePrice = floatPtr(finalPrice)
		item.TotalPrice = finalPrice * float64(item.Quantity)

		cartTotal += item.TotalPrice
	}

	cart.CartTotalPrice = cartTotal
	return cart
}

func buildProductOfferMap(offers []Offer) map[int64][]Offer {
	m := make(map[int64][]Offer)

	for _, offer := range offers {
		if offer.Scope != "product" {
			continue
		}

		for _, id := range offer.EligibleIds {
			m[id] = append(m[id], offer)
		}
	}

	return m
}

// Build categoryID -> []Offer map for O(1) lookups
func buildCategoryOfferMap(offers []Offer) map[int64][]Offer {
	m := make(map[int64][]Offer)

	for _, offer := range offers {
		if offer.Scope != "category" {
			continue
		}

		for _, id := range offer.EligibleIds {
			m[id] = append(m[id], offer)
		}
	}

	return m
}

// Gather both product + category offers for an item
func gatherOffers(
	item *CartItem,
	productMap map[int64][]Offer,
	categoryMap map[int64][]Offer,
) []Offer {

	var result []Offer

	if po, ok := productMap[item.ProductID]; ok {
		result = append(result, po...)
	}

	if co, ok := categoryMap[item.CategoryID]; ok {
		result = append(result, co...)
	}

	return result
}

// Check if offer is active and inside date range
func isOfferValid(o Offer, now time.Time) bool {
	if !o.IsActive {
		return false
	}

	if now.Before(o.StartDate) || now.After(o.EndDate) {
		return false
	}

	return true
}

// Returns discount amount (NOT final price)
// Makes comparison easier
func calculateDiscountAmount(price float64, offer Offer) float64 {
	switch offer.DiscountType {

	case "percentage":
		return price * (offer.DiscountValue / 100)

	case "flat":
		return offer.DiscountValue
	}

	return 0
}

// Small helper to assign *float64
func floatPtr(v float64) *float64 {
	return &v
}
