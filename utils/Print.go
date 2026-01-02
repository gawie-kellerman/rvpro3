package utils

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

var Print printClass

var labelWidth int = 25
var Features uint64 = math.MaxUint64
var PrintDate bool = true
var indent int
var out io.Writer = os.Stdout

type printClass struct{}

func (printClass) Indent(spaces int) {
	indent += spaces
	if indent < 0 {
		indent = 0
	}
}

func (printClass) Detail(label string, format string, a ...any) (n int, err error) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, "%*s", indent, "")
	_, _ = fmt.Fprintf(out, "%-*s", labelWidth-2, label+": ")
	return fmt.Fprintf(out, format, a...)
}

func (printClass) WarnLn(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, "Warn: %*s", indent, "")
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) ErrorLn(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, "Error: %*s", indent, "")
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) InfoLn(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, "Info: %*s", indent, "")
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) Ln(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) RawLn(a ...any) {
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) Fmt(format string, a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, format, a...)
}

func (printClass) Feature(feature int, a ...any) {
	if Print.IsFeature(feature) {
		Print.DatetimeMS(time.Now(), false)
		_, _ = fmt.Fprint(out, a...)
	}
}

func (printClass) FmtFeature(feature int, format string, a ...any) {
	if Print.IsFeature(feature) {
		Print.DatetimeMS(time.Now(), false)
		_, _ = fmt.Fprintf(out, format, a...)
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
		_, _ = fmt.Fprint(out, now.Format(DisplayDateTimeMS), " ")
	}
}
