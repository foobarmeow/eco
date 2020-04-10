package main

import (
	"flag"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"math/rand"
	"os"
	"sort"
	"time"
)

var numAgents int
var numSuppliers int
var intervalCount int
var debugLog bool
var top int
var stagnantMarketCount int
var suppressTables bool

type Position struct {
	X int
	Y int
}

func main() {
	start := time.Now()

	rand.Seed(time.Now().Unix())

	flag.IntVar(&numAgents, "a", 100, "number of agents")
	flag.IntVar(&numSuppliers, "s", 10, "number of suppliers")
	flag.IntVar(&intervalCount, "i", 10, "number of intervals to run")
	flag.IntVar(&top, "t", 5, "top agents")
	flag.BoolVar(&debugLog, "v", false, "show debug logs")
	flag.BoolVar(&suppressTables, "shh", false, "suppress tables")
	flag.Parse()

	// Setup Results table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Transactions", "Quantity", "Revenue", "Unmet Demand"})

	agents := seedAgents()

	// For now this is an invariant, but it shouldn't be,
	// since who the suppliers should theoretically
	// change over intervals.
	suppliers := []Agent{}
	for _, a := range agents {
		if len(a.SupplyMap) > 0 {
			suppliers = append(suppliers, a)
		}
	}

	intervals := 0
	for {
		if intervals == intervalCount {
			break
		}
		intervals++

		fmtDebug("Running interval %d\n-----------------------\n", intervals)

		interval := Interval{Index: intervals}

		for i := range agents {
			agents[i].Purchase(&interval, suppliers)
			agents[i].SeekEmployment(&interval, suppliers)
			agents[i].Reset()
		}

		//fmtDebug("Purchase Phase %d\n-----------------------\n", intervals)
		//for i := range agents {
		//	agents[i].Purchase(&interval, suppliers)
		//}

		//fmtDebug("Work Phase %d\n-----------------------\n", intervals)
		//for i := range agents {
		//	agents[i].SeekEmployment(&interval, suppliers)
		//}

		//fmtDebug("Resetting Demand %d\n-----------------------\n", intervals)
		//for i := range agents {
		//	agents[i].Reset()
		//}

		stats := interval.Stats()

		if stats.IsZero() {
			stagnantMarketCount++
			if stagnantMarketCount > 5 {
				log("STAGNANT MARKET IN", intervals)
				break
			}

		}

		table.Append(stats.Record())
	}

	// Setup Settings Table
	settingsTable := tablewriter.NewWriter(os.Stdout)
	settingsTable.SetHeader([]string{"Intervals", "Agents", "Suppliers"})
	settingsTable.Append([]string{
		fmt.Sprintf("%d", intervalCount),
		fmt.Sprintf("%d", numAgents),
		fmt.Sprintf("%d", numSuppliers),
	})

	// Render agent table
	agentTable := tablewriter.NewWriter(os.Stdout)
	agentTable.SetHeader([]string{"Name", "Cash"})

	sort.Slice(agents, func(a, b int) bool { return agents[a].Cash > agents[b].Cash })

	topAgents := 0
	for _, a := range agents {
		if topAgents == top {
			break
		}
		topAgents++
		agentTable.Append([]string{
			fmt.Sprintf("%s", a.Name),
			fmt.Sprintf("%.2f", a.Cash),
		})
	}

	if !suppressTables {
		table.Render()
		settingsTable.Render()
		agentTable.Render()
	}

	if stagnantMarketCount > 5 {
		fmtDebug("\n Sim took %s to become stagnant in %d intervals \n", time.Since(start), intervals)
		return
	}
	fmtDebug("\n Sim took %s to complete %d intervals \n", time.Since(start), intervals)
}

func seedAgents() []Agent {
	agents := []Agent{}

	if numSuppliers > numAgents {
		numSuppliers = numAgents - 1
	}

	suppliers := numSuppliers

	for {
		if suppliers != 0 {
			agents = append(agents, NewSupplier())
			suppliers--
			continue
		}
		break
	}

	for {
		if numAgents+suppliers == len(agents) {
			break
		}

		agents = append(agents, NewConsumer())
	}

	return agents
}

func fmtDebug(format string, args ...interface{}) {
	if debugLog {
		fmt.Printf(format, args...)
	}
}

func log(args ...interface{}) {
	fmt.Println(args...)
}
