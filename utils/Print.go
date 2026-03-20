package utils

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)
import (
	rt "runtime/debug"
)

var Print printClass
var Test printClass

var labelWidth int = 25
var Features uint64 = math.MaxUint64
var PrintDate bool = true
var indent int
var out io.Writer = os.Stdout

type printClass struct{}

func (printClass) SetIndent(spaces int) {
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
	Print.DateTime()

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("172")).
		MarginRight(1).
		Bold(true)

	fmt.Print(style.Render("WRN"))

	style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("172")).
		Bold(true)

	data := fmt.Sprint(a...)
	fmt.Println(style.Render(data))
}

func (printClass) ErrorLn(a ...any) {
	Print.DateTime()

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("1")).
		MarginRight(1).
		Bold(true)

	fmt.Print(style.Render("ERR"))

	style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Bold(true)

	data := fmt.Sprint(a...)
	fmt.Println(style.Render(data))
}

func (printClass) Stack() {
	stack := rt.Stack()
	stackLines := strings.Split(string(stack), "\n")

	for _, line := range stackLines {
		Print.DateTime()
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("1")).
			MarginRight(1).
			Bold(true)

		fmt.Print(style.Render("STK"))

		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

		data := fmt.Sprint(line)
		fmt.Println(style.Render(data))

	}

}

func (printClass) InfoLn(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	_, _ = fmt.Fprintf(out, "Info: %*s", indent, "")
	_, _ = fmt.Fprintln(out, a...)
}

func (printClass) Ln(a ...any) {
	Print.DateTime()
	//style := lipgloss.NewStyle().
	//	Foreground(lipgloss.Color("14"))

	fmt.Println(a...)
}

func (printClass) Title(a ...interface{}) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	Print.DateTime()

	data := fmt.Sprint(a...)
	fmt.Println(style.Render(data))
}

func (printClass) DateTime() {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		MarginRight(1)

	fmt.Print(style.Render(fmt.Sprint(time.Now().Format(DisplayDateTimeMS))))
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

func (printClass) Indent(ind int) {
	_, _ = fmt.Fprintf(out, "%*s", ind, "")
}

func (printClass) Option(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	Print.Indent(4)
	_, _ = fmt.Println(a...)
}

func (printClass) Descrp(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	Print.Indent(8)
	_, _ = fmt.Println(a...)
}

func (c printClass) Sample(a ...any) {
	Print.DatetimeMS(time.Now(), false)
	Print.Indent(8)
	_, _ = fmt.Println(a...)
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

func (printClass) Command(desc string) {
	Print.DateTime()
	fmt.Println(fmt.Sprintf("  %s", desc))
}

func (printClass) CommandName(name string, desc string) {
	Print.DateTime()

	nameStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("6")).
		MarginLeft(4).Width(15)

	fmt.Print(nameStyle.Render(name))
	fmt.Print(desc)
	fmt.Println()
}

func (printClass) Usage(message string) {
	Print.Ln(message)
}

func (printClass) UsageExample(message string) {
	Print.DateTime()

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		MarginLeft(2)

	fmt.Println(style.Render(message))
}
