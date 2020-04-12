package lib

import (
	"eco/consumable"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestWrite(t *testing.T) {
	// Setup a Market2 with an InventoryMap
	m := NewMarket()

	expectedInventory := Inventory{
		Supply: consumable.Supply{
			Consumable: consumable.NewApple(),
			Quantity:   10,
			Price:      0.25,
		},
	}

	inventories := []Inventory{}
	for i := 0; i < 100; i++ {
		e := expectedInventory
		e.Supply.Quantity += i
		inventories = append(inventories, e)
	}

	m.inventoryMap[consumable.KeyApple] = inventories

	m.Write(consumable.KeyApple, []Inventory{expectedInventory})
	rc := m.Read(consumable.KeyApple)
	for inv := range rc {
		assert.Equal(t, expectedInventory, inv)
	}
}

func TestRead(t *testing.T) {
	// Setup a Market2 with an InventoryMap
	m := NewMarket()

	expectedInventory := Inventory{
		Supply: consumable.Supply{
			Consumable: consumable.NewApple(),
			Quantity:   10,
			Price:      0.25,
		},
	}

	inventories := []Inventory{}
	for i := 0; i < 100; i++ {
		e := expectedInventory
		e.Supply.Quantity += i
		inventories = append(inventories, e)
	}

	m.inventoryMap[consumable.KeyApple] = inventories

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		counter := 0
		rc := m.Read(consumable.KeyApple)
		for inv := range rc {
			e := expectedInventory
			e.Supply.Quantity += counter
			assert.Equal(t, e, inv)
			counter++
		}
		wg.Done()
	}()

	go func() {
		counter := 0
		rc := m.Read(consumable.KeyApple)
		for inv := range rc {
			e := expectedInventory
			e.Supply.Quantity += counter
			assert.Equal(t, e, inv)
			counter++
		}
		wg.Done()
	}()
	wg.Wait()
}
