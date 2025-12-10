package port

type PortIdentifier uint32

const PiObjectList = 88
const PiStatistics = 25
const PiDiagnostics = 86
const PiWgs84 = 137
const PiUncertainty = 157
const PiInstruction = 46
const PiEventTrigger = 24
const PiPVR = 29

func (id PortIdentifier) String() string {
	switch id {
	case PiObjectList:
		return "ObjectList"
	case PiStatistics:
		return "Statistics"
	case PiDiagnostics:
		return "Diagnostics"
	case PiWgs84:
		return "WGS84"
	case PiUncertainty:
		return "Uncertainty"
	case PiInstruction:
		return "instruction"
	case PiEventTrigger:
		return "EventTrigger"
	case PiPVR:
		return "PVR"
	default:
		return "Unknown"
	}
}
