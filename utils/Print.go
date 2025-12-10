package utils

import (
	"fmt"
	"math"
	"os"
	"time"
)

var labelWidth int
var Print printClass
var Features uint64
var PrintDate bool
var indent int

func init() {
	Features = math.MaxUint64
	PrintDate = true
	labelWidth = 25
}

type printClass struct{}

func (printClass) Indent(spaces int) {
	indent += spaces
	if indent < 0 {
		indent = 0
	}
}

func (printClass) Detail(label string, format string, a ...any) (n int, err error) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(os.Stdout, "%*s", indent, "")
	_, _ = fmt.Fprintf(os.Stdout, "%-*s", labelWidth-2, label+": ")
	return fmt.Fprintf(os.Stdout, format, a...)
}

func (printClass) WarnLn(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(os.Stdout, "%*s", indent, "")
	_, _ = fmt.Fprintln(os.Stdout, a...)
}

func (printClass) Ln(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	fmt.Println(a...)
}

func (printClass) RawLn(a ...any) {
	fmt.Println(a...)
}

func (printClass) Fmt(format string, a ...any) {
	Print.DatetimeMS(time.Now(), false)
	fmt.Printf(format, a...)
}

func (printClass) Feature(feature int, a ...any) {
	if Print.IsFeature(feature) {
		Print.DatetimeMS(time.Now(), false)
		fmt.Print(a...)
	}
}

func (printClass) FmtFeature(feature int, format string, a ...any) {
	if Print.IsFeature(feature) {
		Print.DatetimeMS(time.Now(), false)
		fmt.Printf(format, a...)
	}
}

func (printClass) IsFeature(feature int) bool {
	return Features&(1<<uint(feature)) != 0
}

func (printClass) SetFeature(feature int) {
	Features |= 1 << uint(feature)
}

func (printClass) ClearFeature(feature int) {
	Features &= ^(1 << uint(feature))
}

func (printClass) DatetimeMS(now time.Time, forcePrint bool) {
	if PrintDate || forcePrint {
		fmt.Print(Time.ToDisplayDTMS(now), " ")
	}
}
