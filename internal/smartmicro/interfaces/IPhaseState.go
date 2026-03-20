package interfaces

import "rvpro3/radarvision.com/utils"

type IPhaseState interface {
	EqualsRYG(other IPhaseState) bool
	SetRYG(source string, red utils.Uint64, yellow utils.Uint64, green utils.Uint64)
	GetRYG() (utils.Uint64, utils.Uint64, utils.Uint64)
	IsEverSet() bool
}
