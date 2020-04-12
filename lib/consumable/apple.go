package consumable

const KeyApple = "apple"

type Apple interface {
	Key() string
	Scale() int
	Clone() Consumable
	Value() float64
}

type apple struct {
	key   string
	scale int
	value float64
}

func NewApple() Apple {
	return &apple{
		key:   KeyApple,
		value: .25,
		scale: 0,
	}
}

func (a *apple) Key() string {
	return a.key
}

func (a *apple) Scale() int {
	return a.scale
}

func (a *apple) Value() float64 {
	return a.value
}

func (a *apple) Clone() Consumable {
	return &apple{
		key:   KeyApple,
		scale: a.scale,
		value: a.value,
	}
}
