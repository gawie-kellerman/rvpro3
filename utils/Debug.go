package utils

import (
	"os"
)

type debug struct{}

var Debug debug

func init() {
	PrintDate = true
}

func (debug) Panic(err error) {
	if err != nil {
		panic(err)
	}
}

func (debug) AbortWithMsg(msg string, err error) {
	Print.Stack()
	Print.ErrorLn(msg, err)
	os.Exit(1)
}

func (d debug) PanicIf(condition bool, message string) {
	if condition {
		panic(message)
	}
}
