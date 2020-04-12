package lib

import (
	"eco/lib/consumable"
)

type Inventory struct {
	Originator         string
	Price              float64
	Cost               float64
	Goods              []consumable.Consumable
	Consumable         consumable.Consumable
	TransactionChannel chan Transaction
}
