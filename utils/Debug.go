package utils

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
