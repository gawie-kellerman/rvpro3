package interfaces

type IPhaseState interface {
	EqualsRYG(other IPhaseState) bool
	SetRYG(red uint64, yellow uint64, green uint64)
	GetRYG() (uint64, uint64, uint64)
}
