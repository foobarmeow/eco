package lib

import (
	"fmt"
)

var Debug bool
var Verbose bool

func Log(args ...interface{}) {
	log(args)
}

func log(args ...interface{}) {
	if Debug {
		fmt.Println(args...)
	}
}

func fmtDebug(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf(format, args...)
	}
}
