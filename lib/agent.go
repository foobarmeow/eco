package lib

import (
	"eco/lib/consumable"
	"eco/lib/producer"
	"fmt"
	"sync"
	"time"
)

// Agent is the primary struct responsible for acting on the market.
type Agent struct {
	Name string
	Cash float64

	Greed int

	Producers      []producer.Producer
	Consumables    []consumable.Consumable
	Inventory      map[string]Inventory
	Demands        []consumable.Demand
	LaborContracts []LaborContract

	Market             *Market
	LaborMarket        *LaborMarket
	TransactionChannel chan Transaction

	SeeksWage        bool
	IsEmployed       bool
	EmploymentSought bool

	Report Report

	quit   chan bool
	done   chan bool
	rwLock sync.Mutex
}

// NewAgent returns an Agent
func NewAgent(m *Market, l *LaborMarket) Agent {
	return Agent{
		Market:             m,
		LaborMarket:        l,
		TransactionChannel: make(chan Transaction),
		Consumables:        []consumable.Consumable{},
		quit:               make(chan bool),
		done:               make(chan bool),
		rwLock:             sync.Mutex{},
	}
}

func (a *Agent) Start() {
	go a.ProcessTransactions()
}

func (a *Agent) Quit() {
	<-a.done
	a.quit <- true
}

func (a *Agent) ReceiveTick(t time.Time) {
	a.Actions()
}

func (a *Agent) SendToMarket() {
	for k, inv := range a.Inventory {
		price := inv.Cost / float64(len(inv.Goods))
		price += float64(a.Greed) * inv.Consumable.Value()

		a.Report.SentToMarket += len(inv.Goods)

		a.Market.Push(k, Inventory{
			Originator:         a.Name,
			Goods:              inv.Goods,
			Price:              price,
			Consumable:         inv.Consumable,
			TransactionChannel: a.TransactionChannel,
		})
		delete(a.Inventory, k)
		fmtDebug("%s sent %d %s to market.\n", a.Name, len(inv.Goods), k)
	}
}

func (a *Agent) Actions() {
	// lock and get an image of our cash
	a.rwLock.Lock()
	cash := a.Cash
	isEmployed := a.IsEmployed
	a.rwLock.Unlock()
	a.FillDemands(cash)

	if !a.SeeksWage {
		a.SeekLabor()
		a.Produce(cash)
		a.SendToMarket()
	}

	if a.SeeksWage && !isEmployed && !a.EmploymentSought {
		fmtDebug("%s seeks employment.\n", a.Name)
		a.LaborMarket.Append(LaborContract{
			Agent: a,
		})
		a.EmploymentSought = true
	}

}

func (a *Agent) FillDemands(cash float64) {
	for i := range a.Demands {
		d := a.Demands[i]
		a.Market.OrderChannel <- Order{
			From:               a.Name,
			Quantity:           d.Quantity,
			Consumable:         d.Consumable,
			Cash:               cash,
			FulfillmentChannel: a.TransactionChannel,
		}
	}
}

func (a *Agent) SeekLabor() {
	lc := a.LaborMarket.Shift()
	l, hadLabor := <-lc
	if !hadLabor {
		fmtDebug("%s sought labor but there was none.\n", a.Name)
		return
	}

	a.LaborContracts = append(a.LaborContracts, l)
	a.Report.Hired++
	t := true
	l.Agent.TransactionChannel <- Transaction{
		Employment: &t,
		Memo:       fmt.Sprintf("%s has hired %s.", a.Name, l.Agent.Name),
	}
}

func (a *Agent) Produce(cash float64) {
	if len(a.LaborContracts) < 1 {
		fmtDebug("%s has no labor.\n", a.Name)
		return
	}

	for i := range a.Producers {
		p := a.Producers[i]
		fmtDebug("%s will attempt %d production cycles with %s. %.2f\n", a.Name, len(a.LaborContracts), p.Key(), cash)

		estimate := p.Estimate()
		if estimate > cash {
			// We can't produce one cylce,
			// let alone many
			continue
		}

		rate := p.Rate()
		productKey := p.Type().Key()

		totalProduced := 0
		totalWages := 0.0
		totalCost := 0.0

		// Just use up all our labor contracts on the first producer, for now
		for j := range a.LaborContracts {
			cost, wages, products := p.Produce()

			if wages+cost > cash {
				// TODO: What should happen in this situation
				fmtDebug("%s could not afford production cost %.2f (%.2f + %.2f) / %.2f.\n", a.Name, wages+cost, wages, cost, cash)
				continue
			}

			// Deduct the costs
			t := Transaction{
				CashOut:          wages + cost,
				Memo:             fmt.Sprintf("Cost to produce %d %v", rate, productKey),
				From:             p.Key(),
				AcceptChannel:    make(chan bool),
				ResponseRequired: true,
			}
			go func() { a.TransactionChannel <- t }()
			accepted := <-t.AcceptChannel
			if !accepted {
				// can't pay wages
				// TODO: worker should be made aware somehow
				continue
			}

			totalProduced += len(products)
			totalWages += wages
			totalCost += wages + cost

			a.Report.ProductCylces++
			a.Report.Production += rate

			a.Report.WagesPaid += wages

			// Pay the worker
			a.LaborContracts[j].Agent.ReceiveCash(wages, fmt.Sprintf("Wages for producing %d %v", rate, productKey), a.Name)

			inventory, ok := a.Inventory[productKey]
			if !ok {
				inventory = Inventory{Consumable: p.Type(), TransactionChannel: a.TransactionChannel}
			}

			if inventory.TransactionChannel == nil {
				panic("nil trans inv")
			}

			inventory.Cost += wages + cost
			inventory.Goods = append(inventory.Goods, products...)

			a.Inventory[productKey] = inventory
		}
		if totalProduced > 0 {
			fmtDebug("%s produced %d %s and paid %.2f in costs.\n", a.Name, totalProduced, p.Type().Key(), totalCost)
		}
	}
}

func (a *Agent) SendGoods(goods []consumable.Consumable, memo string, from string) {
	a.TransactionChannel <- Transaction{
		ConsumablesIn: goods,
		Memo:          memo,
		From:          from,
	}
}

func (a *Agent) ReceiveCash(amount float64, memo string, from string) {
	a.TransactionChannel <- Transaction{
		CashIn: amount,
		Memo:   memo,
		From:   from,
	}
}

func (a *Agent) DeductCash(cost float64, memo string, from string) {
}

func (a *Agent) ProcessTransactions() {
	for {
		select {
		case t := <-a.TransactionChannel:
			func() {
				a.rwLock.Lock()
				transactionAccepted := true
				defer func() {
					a.rwLock.Unlock()
					if t.ResponseRequired {
						t.AcceptChannel <- transactionAccepted
					}
				}()

				if t.Employment != nil {
					a.IsEmployed = *t.Employment
					fmtDebug("(Employment Change): %s\n", t.Memo)
					return
				}

				qStr := ""
				prefix := ""
				preposition := "from"
				if t.OrderIndex != 0 {
					prefix = fmt.Sprintf("%d: %s received", t.OrderIndex, a.Name)
				}

				if t.CashIn > 0.0 {
					a.Cash += t.CashIn
					a.Report.Revenue += t.CashIn
					prefix = "Received"
					qStr = fmt.Sprintf("%.2f", t.CashIn)
				}

				if t.CashOut > 0.0 {
					if t.CashOut > a.Cash {
						//panic(fmt.Sprintf("%.2f / %.2f %d %s", t.CashOut, a.Cash, t.OrderIndex, a.Name))
						transactionAccepted = false
						return
					}

					a.Cash -= t.CashOut
					preposition = "to"
					qStr = fmt.Sprintf("%.2f", t.CashOut)
					prefix = fmt.Sprintf("%s paid", a.Name)
				}

				suffix := ""
				if len(t.ConsumablesIn) > 0 {
					prefix = fmt.Sprintf("%s paid %.2f to %s for", a.Name, t.CashOut, t.From)
					qStr = fmt.Sprintf("%d %s", len(t.ConsumablesIn), t.ConsumableKey)
					suffix = ""
					a.Consumables = append(a.Consumables, t.ConsumablesIn...)
				}

				if t.Memo != "" {
					suffix = fmt.Sprintf("- (%s)", t.Memo)
				}

				fmtDebug("%s %s %s %s %s\n", prefix, qStr, preposition, t.From, suffix)
			}()

		case <-a.quit:
			return
		}
	}
}

func (a *Agent) ReportRecord() []string {
	//defer func() {
	//	a.Report = Report{}
	//}()
	return []string{
		a.Name,
		fmt.Sprintf("%d", a.Greed),
		fmt.Sprintf("%.2f", a.Cash),
		fmt.Sprintf("%d", len(a.Consumables)),
		fmt.Sprintf("%d", a.Report.SentToMarket),
		fmt.Sprintf("%d", a.Report.Production),
		fmt.Sprintf("%.2f", a.Report.Revenue),
	}
}
