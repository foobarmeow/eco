package lib

type Report struct {
	Wealth        float64
	Consumables   int
	WagesMade     float64
	UnmetDemand   int
	Revenue       float64
	Costs         float64
	SentToMarket  int
	WagesPaid     float64
	Production    int
	ProductCylces int
	Employees     int
	Hired         int
	Fired         int
}

type MarketReport struct {
	TotalCashFlow    float64
	TotalProductFlow int
	ProductReceived  int
	ProductSold      int
	AveragePrice     float64
}
