package interfaces

import "time"

const PhaseStateName = "Phase.State"

type PhaseState struct {
	Red         uint64
	Yellow      uint64
	Green       uint64
	UpdateCount uint64
	UpdateOn    time.Time
	Err         error
}

func (p *PhaseState) EqualsRYG(other IPhaseState) bool {
	if other == nil {
		return false
	}

	r, y, g := other.GetRYG()

	return p.Red == r && p.Yellow == y && p.Green == g
}

func (p *PhaseState) SetRYG(red uint64, yellow uint64, green uint64) {
	p.Red = red
	p.Yellow = yellow
	p.Green = green
	p.UpdateOn = time.Now()
	p.UpdateCount++
}

func (p *PhaseState) GetRYG() (uint64, uint64, uint64) {
	return p.Red, p.Yellow, p.Green
}
