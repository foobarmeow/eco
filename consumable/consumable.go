package consumable

import (
	"math/rand"
)

type Supply struct {
	Key        string
	Quantity   int
	Consumable Consumable
}

type Inventory interface {
	Price() float64
	Quantity() int
}

// Yeesh this bad boy is gonna need generics huh
type Consumable interface {
	// Produce returns an int representing
	// how much of this consumeable is
	// produced through the labor of
	// on Agent over a single Interval
	Produce() int

	// Wage returns the wage for an agent
	// that produces N Product over one Interval.
	Wage() float64

	// Cost returns the cost to produce one unit.
	Cost() float64

	// Price returns the price of one unit
	Price() float64

	// Key returns a string for use in maps
	Key() string

	// Scale returns an int indicating how
	// many are expected to be a "normal"
	// amount that is used in conjunction with
	// a Demand's priority
	Scale() int
}

func RandomConsumable() Consumable {
	switch rand.Intn(2-1) + 1 {
	case 1:
		return NewApple()
	}
	return NewApple()
}
