// Producer is repsonsible for producing a good by consuming labor.
package producer

import (
	"eco/lib/consumable"
)

type Producer interface {
	// Produce returns an int representing
	// how much of a consumable is
	// produced through the labor of
	// on Agent over a single Interval
	Rate() int

	Type() consumable.Consumable

	// Wage returns the wage for an agent
	// that produces N Product over one Interval.
	Wage() float64

	// Cost returns the cost to produce one unit.
	Cost() float64

	// Value returns the cost of the producer.
	Value() float64

	// Key returns a string for use in maps
	Key() string

	// Produce returns the cost (including wage) and wage of one cycle of this Producer
	Produce() (float64, float64, []consumable.Consumable)

	// Estimate returns an "estimate"
	// (i.e. I haven't implemented anything to estimate)
	// of one cycle.
	Estimate() float64
}
