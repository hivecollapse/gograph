package log

import (
	"fmt"
	"gograph/internal/global"
	"log"
	"os"
)

// All informative output goes to StdErr so that
// we can print clean composable output
var StdErr = log.New(os.Stderr, "", 0)

// Display application result
var Out = fmt.Print
var Outln = fmt.Println
var Outf = fmt.Printf

// Print an important message (always visible)
var Print = StdErr.Print
var Println = StdErr.Println
var Printf = StdErr.Printf

// Print verbose message (need -v)
var Verbose = func(v ...any) {
	if global.Verbose || global.Debug {
		StdErr.Print(v...)
	}
}
var Verboseln = func(v ...any) {
	if global.Verbose || global.Debug {
		StdErr.Println(v...)
	}
}
var Verbosef = func(format string, v ...any) {
	if global.Verbose || global.Debug {
		StdErr.Printf(format, v...)
	}
}

// Print debug message (need -d)
var Debug = func(v ...any) {
	if global.Debug {
		StdErr.Print(v...)
	}
}
var Debugln = func(v ...any) {
	if global.Debug {
		StdErr.Println(v...)
	}
}
var Debugf = func(format string, v ...any) {
	if global.Debug {
		StdErr.Printf(format, v...)
	}
}

// Fatal
var Fatal = StdErr.Fatal
var Fatalln = StdErr.Fatalln
var Fatalf = StdErr.Fatalf

// Panic
var Panic = StdErr.Panic
var Panicln = StdErr.Panicln
var Panicf = StdErr.Panicf
