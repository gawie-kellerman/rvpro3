package interfaces

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

const PhaseStateName = "Phase.State"

type PhaseState struct {
	PhaseRed    utils.Uint64
	PhaseYellow utils.Uint64
	PhaseGreen  utils.Uint64
	UpdateCount uint64
	UpdateOn    time.Time
	ChangeOn    time.Time
	ChangeCount uint64
	Source      string
	Err         error
}

func (p *PhaseState) EqualsRYG(other IPhaseState) bool {
	if other == nil {
		return false
	}

	r, y, g := other.GetRYG()

	return p.PhaseRed == r && p.PhaseYellow == y && p.PhaseGreen == g
}

func (p *PhaseState) SetRYG(source string, red utils.Uint64, yellow utils.Uint64, green utils.Uint64) {
	p.UpdateOn = utils.Time.Exact()
	p.UpdateCount++

	if p.PhaseRed != red || p.PhaseYellow != yellow || p.PhaseGreen != green {
		p.Source = source
		p.PhaseRed = red
		p.PhaseYellow = yellow
		p.PhaseGreen = green
		p.ChangeCount++
		p.ChangeOn = p.UpdateOn
	}
}

func (p *PhaseState) GetRYG() (utils.Uint64, utils.Uint64, utils.Uint64) {
	return p.PhaseRed, p.PhaseYellow, p.PhaseGreen
}

func (p *PhaseState) IsEverSet() bool {
	return p.UpdateCount != 0
}
