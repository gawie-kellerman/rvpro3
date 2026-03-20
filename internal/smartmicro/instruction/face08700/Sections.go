package face08700

const AppTMParametersSection = 3017
const S3017SimulationMode = 3
const S3017SimulationModeName = "simulation_mode"

const TRObjectListSection = 218
const S218SimulationMode = 8

type S218SimulationModeEnum uint32

const (
	SimModeDisabled S218SimulationModeEnum = iota
	SimModeLines
	SimModeSplines
)

func (sm S218SimulationModeEnum) String() string {
	switch sm {
	case SimModeDisabled:
		return "disabled (0)"
	case SimModeLines:
		return "lines (1)"
	case SimModeSplines:
		return "splines (2)"
	default:
		return "unknown"
	}
}
