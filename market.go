package main

import (
	"eco/consumable"
)

type Market struct {
	SupplyMap map[string][]consumable.Inventory
}

func (m *Market) AveragePrice(consumeableKey string) (bool, float64) {
	if inventories, ok := m.SupplyMap[consumeableKey]; ok {
		total := 0.0
		for _, i := range inventories {
			total += i.Price()
		}
		return true, total / float64(len(inventories))
	}

	return false, 0
}
