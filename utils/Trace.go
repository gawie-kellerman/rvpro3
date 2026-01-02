package utils

import "fmt"

// traceClass is used for **TEMPORARY** developer tracing only
// This is purposefully done to avoid fmt.Print* function calls in
// production code, as these forces
// a global find to find (which would return too many results for fmt.Print)
// When using traceClass you can find usages (which is much better)
type traceClass struct{}

var TraceClass traceClass

func (traceClass) Println(args ...any) {
	fmt.Println(args...)
}

func (traceClass) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (traceClass) Print(args ...any) {
	fmt.Print(args...)
}
