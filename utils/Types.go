package utils

import (
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// ParseInt parses a string as an integer
// A string starting with 0x or 0X is considered hexadecimal
// A string starting with 0b or 0B is considered binary
// A string starting with 0o or oO is considered octal
// Otherwise the string will be parsed as decimal
func ParseInt[T constraints.Integer](value string, defValue T) (T, error) {
	if IsHexNumber(value) {
		res, err := strconv.ParseInt(value[2:], 16, 64)

		if err != nil {
			return 0, err
		}

		return T(res), nil
	}

	if IsBinary(value) {
		res, err := strconv.ParseInt(value[2:], 2, 64)
		if err != nil {
			return 0, err
		}

		return T(res), nil
	}

	if IsOctal(value) {
		res, err := strconv.ParseInt(value[2:], 8, 64)
		if err != nil {
			return 0, err
		}
		return T(res), nil
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return T(intVal), nil
}

func ParseFloat[T constraints.Float](value string, defValue T) (T, error) {
	floatVal, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return 0, err
	}
	return T(floatVal), nil
}

func IsHexNumber(value string) bool {
	return strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X")
}

func IsBinary(value string) bool {
	return strings.HasPrefix(value, "0b") || strings.HasPrefix(value, "0B")
}

func IsOctal(value string) bool {
	return strings.HasPrefix(value, "0o") || strings.HasPrefix(value, "0O")
}
