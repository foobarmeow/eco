package producer

import (
	"eco/lib/consumable"
)

const KeyOrchard = "orchard"

type orchard struct {
	rate           int
	productionType consumable.Consumable
	cost           float64
	wage           float64
	key            string
	value          float64
	products       []consumable.Consumable
}

func NewOrchard() Producer {
	return &orchard{
		rate:     10,
		cost:     1,
		wage:     5,
		value:    .25,
		key:      KeyOrchard,
		products: []consumable.Consumable{},
	}
}

func (o *orchard) Rate() int {
	return o.rate
}

func (o *orchard) Wage() float64 {
	return o.wage
}

func (o *orchard) Cost() float64 {
	return o.cost
}

func (o *orchard) Key() string {
	return o.key
}

func (o *orchard) Value() float64 {
	return o.value
}

func (o *orchard) Type() consumable.Consumable {
	return consumable.NewApple()
}

func (o *orchard) Products() []consumable.Consumable {
	p := o.products
	o.products = []consumable.Consumable{}
	return p
}

func (o *orchard) Produce() (float64, float64, []consumable.Consumable) {
	products := []consumable.Consumable{}
	for i := 0; i < o.rate; i++ {
		products = append(products, o.Type().Clone())
	}
	wage := float64(o.rate) * o.wage
	cost := (float64(o.rate) * o.cost) + wage
	return cost, wage, products
}

func (o *orchard) Estimate() float64 {
	// TODO: cache this?
	for i := 0; i < o.rate; i++ {
		o.products = append(o.products, o.Type().Clone())
	}
	wage := float64(o.rate) * o.wage
	return (float64(o.rate) * o.cost) + wage
}
