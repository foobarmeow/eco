package main

import (
	"eco/consumable"
	"github.com/Pallinder/go-randomdata"
	"math"
	"math/rand"
)

type Demand struct {
	Priority   int
	Consumable consumable.Consumable
	Quantity   int
}

type Agent struct {
	Name string

	NeedScale        int
	ExpectationScale int
	Demands          []Demand

	SupplyMap map[string]consumable.Supply

	Cash float64
}

func NewSupplier() Agent {
	// Suppliers don't have demands yet
	a := Agent{
		Name:      randomdata.State(randomdata.Large),
		Cash:      float64(rand.Intn(200-20) + 20),
		SupplyMap: map[string]consumable.Supply{},
	}

	a.SupplyMap[consumable.KeyApple] = consumable.Supply{
		Consumable: consumable.NewApple(),
		// They start with no apples!
		Quantity: 0,
	}

	return a
}

func NewConsumer() Agent {
	a := Agent{
		Name: randomdata.LastName(),
	}

	// The juice, as it were
	a.NeedScale = rand.Intn(20-10) + 10
	a.ExpectationScale = rand.Intn(50-20) + 20

	a.Demands = []Demand{
		{
			Priority:   1,
			Consumable: consumable.NewApple(),
			Quantity:   a.ExpectationScale * a.NeedScale,
		},
	}

	return a
}

func (a *Agent) Reset() {
	if len(a.Demands) < 1 {
		return
	}

	for i := range a.Demands {
		d := a.Demands[i]
		d.Quantity = a.ExpectationScale * a.NeedScale * d.Consumable.Scale()
	}
}

func (a *Agent) ShouldHire(candidate *Agent) (float64, bool) {
	for k, s := range a.SupplyMap {
		// CONSIDER PROFITS
		if a.Cash < 2000 && s.Quantity < 1000 {
			p := s.Consumable.Produce()
			s.Quantity += p
			a.SupplyMap[k] = s
			a.Cash -= s.Consumable.Wage()

			fmtDebug("\tAgent %s employed Agent %s and got %d from their labor! (Q: %d, C: %.2f)\n", a.Name, candidate.Name, p, s.Quantity, a.Cash)

			return s.Consumable.Wage(), true
		}
	}
	return 0.0, false
}

// Sell returns the price of the demanded supplies
// and subtracts the number of supplies demanded
// from the seller's supply
func (a *Agent) Sell(goodType string, quantity int) float64 {
	// Ideally we would have already checked that this demand
	// can be supplied (both the supply exists and in the
	// quantity demanded).
	supply, _ := a.SupplyMap[goodType]
	supply.Quantity = supply.Quantity - quantity
	a.SupplyMap[goodType] = supply

	return float64(quantity) * supply.Consumable.Price()
}

// Purchase attempts to purchase supplies and
// returns the transaction(s) that took place
func (a *Agent) Purchase(interval *Interval, suppliers []Agent) {
	// Loop through the agents demands
	// Attempt to find an agent that can supply them
	// TODO: Maybe demands have to be sorted by priority
	spent := 0.0
	unmetDemand := 0

	defer func() {
		interval.Reports = append(interval.Reports, Report{
			Agent:       *a,
			Spent:       spent,
			UnmetDemand: unmetDemand,
		})
	}()

	for i := range a.Demands {
		d := &a.Demands[i]

		k := d.Consumable.Key()
		fmtDebug("Agent %s wants to buy %d %v with %.2f cash.\n", a.Name, d.Quantity, k, a.Cash)

		for j := range suppliers {
			if d.Quantity == 0 {
				fmtDebug("\tAgent %s passes since their demand for %v is met.\n", a.Name, k)
				continue
			}

			supplier := suppliers[j]

			var supply consumable.Supply
			var ok bool
			if supply, ok = supplier.SupplyMap[d.Consumable.Key()]; !ok {
				// They don't carry it
				continue
			}

			if supply.Quantity == 0 {
				// They carried it, but are out
				fmtDebug("\tAgent %s failed to buy %d %v from Agent %s for lack of supply.\n", a.Name, d.Quantity, k, supplier.Name)
				continue
			}

			quantity := d.Quantity
			if supply.Quantity < d.Quantity {
				// If the supplier has less than we want, just buy that
				quantity = supply.Quantity
			}

			consumable := supply.Consumable

			totalPrice := consumable.Price() * float64(quantity)

			if a.Cash < totalPrice {
				// TODO: Should they reduce how many they want to buy, or try another seller?
				// Can't afford it, move on
				quantity = int(math.Round(a.Cash / d.Consumable.Price()))
				if quantity == 0 {
					fmtDebug("\tAgent %s failed to buy %d %v from Agent %s for lack of cash (total: %.2f, had: %.2f).\n", a.Name, d.Quantity, k, supplier.Name, totalPrice, a.Cash)
					continue
				}
				totalPrice = consumable.Price() * float64(quantity)
				//continue
			}

			transaction := Transaction{
				Demand: *d,
				Quote: Quote{
					Quantity:  quantity,
					UnitPrice: consumable.Price(),
					Price:     totalPrice,
				},
			}

			// Let's purchase!
			a.Cash -= supplier.Sell(k, quantity)
			d.Quantity -= quantity

			spent += totalPrice

			fmtDebug("\tAgent %s bought %d %v with %.2f cash from Agent %s whose supply is now at %d.\n", a.Name, quantity, k, totalPrice, supplier.Name, supply.Quantity-quantity)
			fmtDebug("\t\tAgent %s's demand for %v is now %d.\n", a.Name, k, d.Quantity)

			transaction.Buyer = *a
			transaction.Supplier = supplier
			interval.Transactions = append(interval.Transactions, transaction)

			if a.Cash == 0 {
				return
			}
		}
		unmetDemand += d.Quantity
	}
}

func (a *Agent) SeekEmployment(interval *Interval, employers []Agent) {
	// Determine if this agent needs work
	// If they have demands and a lack of cash,
	// yeah they're gonna wanna work
	if len(a.Demands) == 0 {
		return
	}

	demandScore := 0
	for i := range a.Demands {
		d := &a.Demands[i]
		demandScore += d.Quantity * d.Priority * d.Consumable.Scale()
	}

	if demandScore == 0 {
		fmtDebug("Agent %s's demandScore was %d so they are NOT going to seek work.\n", a.Name, demandScore)
		return
	}

	fmtDebug("Agent %s's demandScore was %d so they are going to seek work.\n", a.Name, demandScore)

	// Shuffle the employers so maybe they all get ot choose
	rand.Shuffle(len(employers), func(i, j int) { employers[i], employers[j] = employers[j], employers[i] })

	// Weight the demand score against their cash?
	// ...

	// .....

	// Okay so they have to work now
	//unemployed := true
	totalTimesTheyCanWork := 1
	for {
		if totalTimesTheyCanWork == 0 {
			break
		}
		totalTimesTheyCanWork--

		for i := range employers {
			e := employers[i]
			var wage float64
			var ok bool
			if wage, ok = e.ShouldHire(a); ok {
				a.Cash += wage
				fmtDebug("\tAgent %s got work with %s and made %.2f!\n", a.Name, e.Name, wage)
			}
		}
	}
}
