package lib

import (
	"sync"
)

type LaborContract struct {
	Agent *Agent
}

type LaborMarket struct {
	labor  []LaborContract
	rwLock sync.Mutex
	quit   chan bool
	done   chan bool
}

func NewLaborMarket() LaborMarket {
	return LaborMarket{
		labor:  []LaborContract{},
		rwLock: sync.Mutex{},
		quit:   make(chan bool),
		done:   make(chan bool),
	}
}

func (m *LaborMarket) Append(contract LaborContract) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	m.labor = append(m.labor, contract)
}

func (m *LaborMarket) Shift() <-chan LaborContract {
	c := make(chan LaborContract)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()
		defer close(c)

		if len(m.labor) < 1 {
			return
		}
		log("b", m.labor)

		l, labor := m.labor[0], m.labor[1:]
		m.labor = labor

		c <- l
	}()

	return c
}

func (m *LaborMarket) Read() <-chan LaborContract {
	c := make(chan LaborContract)
	go func() {
		m.rwLock.Lock()
		defer m.rwLock.Unlock()
		defer close(c)

		for i := range m.labor {
			c <- m.labor[i]
		}
	}()

	return c
}
