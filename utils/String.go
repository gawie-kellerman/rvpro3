package utils

import (
	"strconv"
	"strings"
)

var String stringUtil

type stringUtil struct {
}

func (stringUtil) ToInt64(source string) (int64, error) {
	if strings.HasPrefix(source, "0x") {
		source = source[2:]
		return strconv.ParseInt(source, 16, 32)
	}

	if strings.HasPrefix(source, "0b") {
		source = source[2:]
		return strconv.ParseInt(source, 2, 64)
	}

	if strings.HasPrefix(source, "0o") {
		source = source[2:]
		return strconv.ParseInt(source, 8, 32)
	}

	return strconv.ParseInt(source, 10, 64)
}

func (stringUtil) Or(option1 string, option2 string) string {
	if option1 == "" {
		return option2
	}
	return option1
}
