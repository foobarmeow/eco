package consumable

type Consumable interface {
	// Key returns a string for use in maps
	Key() string

	// Scale returns an int indicating how
	// many are expected to be a "normal"
	// amount that is used in conjunction with
	// a Demand's priority
	Scale() int

	// Value returns the base value of this consumable
	Value() float64

	Clone() Consumable
}
