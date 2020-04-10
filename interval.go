package main

import (
	"fmt"
)

type Stat struct {
	Index           int
	NumTransactions int
	Quantity        int
	CashFlow        float64
	UnmetDemand     int
}

func (s *Stat) IsZero() bool {
	return s.Quantity == 0 && s.CashFlow == 0
}

func (s *Stat) Record() []string {
	return []string{
		fmt.Sprintf("%d", s.Index),
		fmt.Sprintf("%d", s.NumTransactions),
		fmt.Sprintf("%d", s.Quantity),
		fmt.Sprintf("%.2f", s.CashFlow),
		fmt.Sprintf("%d", s.UnmetDemand),
	}
}

type Report struct {
	Agent       Agent
	Spent       float64
	Made        float64
	UnmetDemand int
	UnsoldGoods int
}

type Quote struct {
	Quantity  int
	UnitPrice float64
	Price     float64
}

type Transaction struct {
	Buyer    Agent
	Supplier Agent
	Demand   Demand
	Quote    Quote
}

type Interval struct {
	Index        int
	Transactions []Transaction
	Reports      []Report
}

func (i *Interval) Stats() Stat {
	s := Stat{
		Index:           i.Index,
		NumTransactions: len(i.Transactions),
	}

	for _, t := range i.Transactions {
		s.Quantity += t.Quote.Quantity
		s.CashFlow += t.Quote.Price
	}

	for _, r := range i.Reports {
		s.UnmetDemand += r.UnmetDemand
	}
	return s
}

func (i *Interval) StatsRecord() []string {
	quantity := 0
	cashFlow := 0.0
	unmetDemand := 0

	for _, t := range i.Transactions {
		quantity += t.Quote.Quantity
		cashFlow += t.Quote.Price
	}

	for _, r := range i.Reports {
		unmetDemand += r.UnmetDemand
	}

	return []string{
		fmt.Sprintf("%d", i.Index),
		fmt.Sprintf("%d", len(i.Transactions)),
		fmt.Sprintf("%d", quantity),
		fmt.Sprintf("%.2f", cashFlow),
		fmt.Sprintf("%d", unmetDemand),
	}
}
