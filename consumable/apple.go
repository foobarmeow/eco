package consumable

const KeyApple = "apple"

type Apple interface {
	Produce() int
	Wage() float64
	Cost() float64
	Price() float64
	Key() string
	Scale() int
}

type apple struct {
	production int
	cost       float64
	wage       float64
	price      float64
	key        string
	scale      int
}

func NewApple() Apple {
	return &apple{
		production: 10,
		cost:       1,
		wage:       5,
		price:      2,
		key:        KeyApple,
		scale:      1000,
	}
}

func (a *apple) Produce() int {
	return a.production
}

func (a *apple) Wage() float64 {
	return a.wage
}

func (a *apple) Cost() float64 {
	return a.cost
}

func (a *apple) Price() float64 {
	return a.price
}

func (a *apple) Key() string {
	return a.key
}

func (a *apple) Scale() int {
	return a.scale
}
