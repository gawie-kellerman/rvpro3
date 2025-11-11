package utils

type RetryGuard struct {
	ModCycles uint32
	Cycles    uint32
}

func (r RetryGuard) ShouldRetry() bool {
	r.Cycles++
	if r.ModCycles == 0 {
		r.ModCycles = 5
	}
	return r.Cycles%r.ModCycles == 1
}

func (r RetryGuard) Reset() {
	r.Cycles = 0
}
