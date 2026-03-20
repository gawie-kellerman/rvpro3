package utils

type exec struct{}

var Exec exec

func (exec *exec) If(condition bool, callback func()) {
	if condition {
		callback()
	}
}
