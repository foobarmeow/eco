// Heyo
package main

import (
	"bufio"
	"eco/lib"
	"eco/lib/consumable"
	"eco/lib/producer"
	"eco/lib/ui"
	"flag"
	"github.com/Pallinder/go-randomdata"
	"github.com/olekukonko/tablewriter"
	"math/rand"
	"os"
	//"sort"
	"sync"
	"time"
)

var timeout int
var interval int
var verbose bool
var debug bool
var step bool
var agentCount int
var supplierCount int
var suppressTables bool

func main() {
	rand.Seed(time.Now().Unix())

	flag.IntVar(&interval, "i", 100, "tick interval in ms")
	flag.IntVar(&timeout, "t", 0, "sim timeout")
	flag.IntVar(&agentCount, "ac", 10, "count of agents")
	flag.IntVar(&supplierCount, "sc", 3, "count of suppliers")
	flag.BoolVar(&verbose, "v", false, "print logs")
	flag.BoolVar(&debug, "d", false, "print debug logs")
	flag.BoolVar(&step, "step", false, "step through ticks")
	flag.BoolVar(&suppressTables, "shh", false, "suppressTables")
	flag.Parse()

	lib.Debug = debug
	lib.Verbose = verbose

	m := lib.NewMarket()
	l := lib.NewLaborMarket()

	agents := []lib.Agent{}
	agents = append(agents, GenerateConsumers(agentCount-supplierCount, &m, &l)...)
	agents = append(agents, GenerateSuppliers(supplierCount, &m, &l)...)

	// Start the market
	go m.Start()

	// Startup agents
	for i := range agents {
		go agents[i].Start()
	}

	// Send ticks to each agent
	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(time.Duration(timeout) * time.Second)
	}

	wg := sync.WaitGroup{}

	report := make(chan chan bool)
	reports := make(chan []string)

	graph := ui.NewGraph()

	// Collect agent reports and report them in a table
	go func() {
		records := [][]string{}
		for {
			select {
			case done := <-report:
				// Render Agents table
				agentsTable := tablewriter.NewWriter(os.Stdout)
				agentsTable.SetHeader([]string{"Name", "Greed", "Cash", "Consumables", "Market Sent", "Produced", "Revenue"})
				agentsTable.AppendBulk(records)
				records = [][]string{}

				graph.Update(m.MarketReport())

				// Render Market table
				returnedRecord := make(chan []string)
				m.ReportChannel <- returnedRecord
				r := <-returnedRecord
				marketTable := tablewriter.NewWriter(os.Stdout)
				marketTable.SetHeader([]string{"Sold", "Received", "Total Cash Flow", "Avg Price", "Stock"})
				marketTable.Append(r)

				if !suppressTables {
					agentsTable.Render()
					marketTable.Render()
				}

				done <- true
			case r := <-reports:
				records = append(records, r)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-timeoutChan:
				if step {
					continue
				}
				m.Quit()
				return
			default:
				wg.Add(len(agents))
				for i := range agents {
					a := &agents[i]
					go func() {
						reports <- a.Actions()
						wg.Done()
					}()
				}
				wg.Wait()

				resume := make(chan bool)
				report <- resume
				<-resume

				if step {
					input := bufio.NewScanner(os.Stdin)
					input.Scan()
				}
			}
		}
	}()

	graph.Start()
}

func NewRandomizedConsumer(m *lib.Market, l *lib.LaborMarket) lib.Agent {
	a := lib.NewAgent(m, l)
	a.Name = randomdata.LastName()
	a.SeeksWage = true
	a.Cash = RandomCash()
	a.Demands = []consumable.Demand{
		{
			Consumable: consumable.NewApple(),
			Quantity:   rand.Intn(10000-200) + 200,
		},
	}
	return a
}

func NewRandomizedSupplier(m *lib.Market, l *lib.LaborMarket) lib.Agent {
	a := lib.NewAgent(m, l)
	a.Name = randomdata.State(randomdata.Large)
	a.Cash = RandomCash()
	a.Greed = rand.Intn(200-20) + 20
	a.Producers = append(a.Producers, producer.NewOrchard())
	a.Inventory = map[string]lib.Inventory{}
	return a
}

func RandomCash() float64 {
	return float64(rand.Intn(2000-500) + 500)
}

func GenerateConsumers(count int, m *lib.Market, l *lib.LaborMarket) []lib.Agent {
	agents := []lib.Agent{}
	for i := 0; i < count; i++ {
		agents = append(agents, NewRandomizedConsumer(m, l))
	}
	return agents
}

func GenerateSuppliers(count int, m *lib.Market, l *lib.LaborMarket) []lib.Agent {
	agents := []lib.Agent{}
	for i := 0; i < count; i++ {
		agents = append(agents, NewRandomizedSupplier(m, l))
	}
	return agents
}
