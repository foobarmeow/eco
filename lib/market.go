package lib

import (
	"eco/lib/consumable"
	"fmt"
	"sort"
	"sync"
	"time"
)

type MarketReport struct {
	TotalCashFlow    float64
	TotalProductFlow int
	ProductReceived  int
	ProductSold      int
	AveragePrice     float64
}

// Transaction holds all relevant information
// about a transaction. Employment changes occur
// via transactions as well as all market transactions.
type Transaction struct {
	CashIn         float64
	CashOut        float64
	ConsumablesIn  []consumable.Consumable
	ConsumablesOut []consumable.Consumable

	Employment *bool

	From             string
	Memo             string
	Time             time.Time
	ConsumableKey    string
	OrderIndex       int
	AcceptChannel    chan bool
	ResponseRequired bool
}

// Market coordinates transactions of goods
type Market struct {
	OrderChannel  chan Order
	ReportChannel chan chan []string

	inventoryMap map[string][]Inventory
	rwLock       sync.Mutex
	quit         chan bool
	done         chan bool
	report       MarketReport
}

// NewMarket returns a new Markey
func NewMarket() Market {
	return Market{
		ReportChannel: make(chan chan []string),
		OrderChannel:  make(chan Order, 100),
		inventoryMap:  map[string][]Inventory{},
		rwLock:        sync.Mutex{},
		quit:          make(chan bool),
	}
}

// Quit stops m.ProcessOrders
func (m *Market) Quit() {
	m.quit <- true
}

// Start starts m.ProcessOrders
func (m *Market) Start() {
	m.ProcessOrders()
}

// ProcessOrders processes orders
func (m *Market) ProcessOrders() {
	count := 0
	for {
		select {
		case returnChan := <-m.ReportChannel:
			returnChan <- m.Report()
			continue
		case order := <-m.OrderChannel:
			count++

			order.Index = count

			name := order.From
			quantity := order.Quantity
			key := order.Consumable.Key()

			fmtDebug("Order %d: Market received order from %s for %d %s. %.2f\n", count, name, quantity, key, order.Cash)

			// Check to see if they can even afford one unit at the lowest price
			func() {
				lowest := <-m.ReadLowest(key)
				if len(lowest.Goods) == 0 {
					// Can't fill this order
					fmtDebug("\tOrder %d: Market has no inventory of %s for %s.\n", count, key, name)
					return
				}

				if order.Cash < lowest.Price {
					fmtDebug("\tOrder %d: %s couldn't afford any units of %s at %.2f. (%.2f, 1)\n", count, name, key, lowest.Price, order.Cash)
					return
				}

				if q := order.PurchasableQuantity(lowest); q < 1 {
					fmtDebug("\tOrder %d: %s couldn't afford any units of %s at %.2f. (%.2f, 2)\n", count, name, key, lowest.Price, order.Cash)
					return
				} else {
					quantity = q
				}

				// At this point we want to purchase it
				// Do so synchronously
				// Give the purchaser a chance to decline
				invChan, confirm := m.PopConfirm(key)
				inv := <-invChan
				if len(inv.Goods) == 0 {
					fmtDebug("\tOrder %d: no supply\n", count)
					confirm <- false
					return
				}

				t := Transaction{
					ConsumableKey:    key,
					ConsumablesIn:    inv.Goods,
					CashOut:          float64(quantity) * inv.Price,
					From:             inv.Originator,
					OrderIndex:       order.Index,
					AcceptChannel:    make(chan bool),
					ResponseRequired: true,
				}

				go func() { order.FulfillmentChannel <- t }()
				accepted := <-t.AcceptChannel
				if !accepted {
					fmtDebug("\tOrder %d: not accepted\n", count)
					confirm <- false
					return
				}

				confirm <- true

				price := float64(len(inv.Goods)) * inv.Price

				m.report.ProductSold += len(inv.Goods)
				m.report.TotalCashFlow += price

				// Send money to originator
				inv.TransactionChannel <- Transaction{
					CashIn:     price,
					From:       order.From,
					OrderIndex: order.Index,
				}
				fmtDebug("\tOrder %d: order filled %d\n", order.Index, len(inv.Goods))
			}()
		case <-m.quit:
			return
		}
	}
}

// Push appends the inventory to the slice at key
func (m *Market) Push(key string, inv Inventory) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	inventories, ok := m.inventoryMap[key]
	if !ok {
		m.inventoryMap[key] = []Inventory{inv}
		return
	}

	inventories = append(inventories, inv)

	m.report.ProductReceived += len(inv.Goods)

	// Sort the inventories by price
	// TODO: This is potentially expensive
	sort.Slice(inventories, func(i, j int) bool {
		return inventories[i].Price < inventories[j].Price
	})

	m.inventoryMap[key] = inventories
}

// ReadLowest returns a channel that returns the lowest priced
// inventory from the slice at key
func (m *Market) ReadLowest(key string) <-chan Inventory {
	c := make(chan Inventory)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()

		inventories, ok := m.inventoryMap[key]
		if !ok || len(inventories) < 1 {
			go func() { c <- Inventory{} }()
			return
		}

		defer close(c)
		c <- inventories[0]
	}()

	return c
}

func (m *Market) Pop(key string) <-chan Inventory {
	c := make(chan Inventory)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()

		var inventory Inventory
		inventories, ok := m.inventoryMap[key]
		if !ok || len(inventories) < 1 {
			go func() { c <- Inventory{} }()
			return
		}

		inventory, inventories = inventories[0], inventories[1:]
		m.inventoryMap[key] = inventories

		defer close(c)
		c <- inventory
	}()

	return c
}

func (m *Market) PopConfirm(key string) (<-chan Inventory, chan bool) {
	c := make(chan Inventory)
	confirm := make(chan bool)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()
		defer close(c)

		var inventory Inventory
		inventories, ok := m.inventoryMap[key]
		if !ok || len(inventories) < 1 {
			go func() { c <- Inventory{} }()
			return
		}

		inventory, inventories = inventories[0], inventories[1:]

		go func() {
			c <- inventory
		}()

		if <-confirm {
			m.inventoryMap[key] = inventories
		}
	}()

	return c, confirm
}

// Read returns a channel that returns all inventories in the slice at key
func (m *Market) Read(key string) <-chan Inventory {
	c := make(chan Inventory, 1)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()

		inventories, ok := m.inventoryMap[key]
		if !ok {
			go func() { c <- Inventory{} }()
			return
		}

		defer close(c)
		for _, i := range inventories {
			c <- i
		}
	}()

	return c
}

func (m *Market) Report() []string {
	m.rwLock.Lock()
	defer func() {
		m.report = MarketReport{}
		m.rwLock.Unlock()
	}()

	stock := 0
	for _, inventories := range m.inventoryMap {
		for _, inv := range inventories {
			stock = len(inv.Goods)
		}
	}

	avg := 0.0
	if m.report.TotalCashFlow > 0.0 {
		avg = m.report.TotalCashFlow / float64(m.report.ProductSold)
	}
	return []string{
		fmt.Sprintf("%d", m.report.ProductSold),
		fmt.Sprintf("%d", m.report.ProductReceived),
		fmt.Sprintf("%.2f", m.report.TotalCashFlow),
		fmt.Sprintf("%.2f", avg),
		fmt.Sprintf("%d", stock),
	}
}
