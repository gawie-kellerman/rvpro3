package utils

type RetryGuard struct {
	RetryEvery uint32
	Cycles     uint32
}

func (r *RetryGuard) ShouldRetry() bool {
	res := r.Cycles%r.RetryEvery == 0
	r.Cycles++
	//if r.RetryEvery == 0 {
	//	r.RetryEvery = 5
	//}
	return res
}

func (r *RetryGuard) Reset() {
	r.Cycles = 0
}
