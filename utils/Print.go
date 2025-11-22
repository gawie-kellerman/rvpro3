package utils

import (
	"fmt"
	"os"
)

var labelWidth int
var Print printClass
var indent int

func init() {
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
	_, _ = fmt.Fprintf(os.Stdout, "%*s", indent, "")
	_, _ = fmt.Fprintf(os.Stdout, "%-*s", labelWidth-2, label+": ")
	return fmt.Fprintf(os.Stdout, format, a...)
}
