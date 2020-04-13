package lib

import (
	"eco/lib/consumable"
)

type Order struct {
	Index              int
	From               string
	Cash               float64
	Quantity           int
	Consumable         consumable.Consumable
	FulfillmentChannel chan Transaction
}

func (o *Order) CanAffordAt(inv Inventory) bool {
	return o.Cash > float64(o.Quantity)*inv.Price
}

func (o *Order) CanAffordQuantityAt(q int, inv Inventory) bool {
	return o.Cash > float64(q)*inv.Price
}

func (o *Order) PurchasableQuantity(inv Inventory) int {
	// Check to see if they can afford their desired quantity
	// at this price.
	quantity := o.Quantity
	if quantity > len(inv.Goods) {
		quantity = len(inv.Goods)
	}

	if !o.CanAffordQuantityAt(quantity, inv) {
		// Propose that they only buy what they can afford
		div := o.Cash / inv.Price
		if div > 1 {
			return int(div)
		}
		return 0
	}
	return quantity
}
