package triggerpipeline

type ChannelStatus byte

const (
	ChannelStatusUnused      ChannelStatus = ChannelStatus('2')
	ChannelStatusNoCall      ChannelStatus = ChannelStatus('0')
	ChannelStatusCall        ChannelStatus = ChannelStatus('1')
	ChannelStatusRedExtend   ChannelStatus = ChannelStatus('E')
	ChannelStatusRedHold     ChannelStatus = ChannelStatus('H')
	ChannelStatusDilemma     ChannelStatus = ChannelStatus('D')
	ChannelStatusDilemmaHold ChannelStatus = ChannelStatus('d')
	ChannelStatusWatchDog    ChannelStatus = ChannelStatus('W')
	ChannelStatusOpenCircuit ChannelStatus = ChannelStatus('O')
	ChannelStatusShort       ChannelStatus = ChannelStatus('S')
	ChannelStatusFailSafeOff ChannelStatus = ChannelStatus('f')
	ChannelStatusFailSafeOn  ChannelStatus = ChannelStatus('F')
	ChannelStatusForceClear  ChannelStatus = ChannelStatus('C')
	ChannelStatusForceSet    ChannelStatus = ChannelStatus('c')
)

func (status ChannelStatus) String() string {
	switch status {
	case ChannelStatusUnused:
		return "Unused"
	case ChannelStatusNoCall:
		return "NoCall"
	case ChannelStatusCall:
		return "Call"
	case ChannelStatusRedExtend:
		return "RedExtend"
	case ChannelStatusRedHold:

		return "RedHold"
	case ChannelStatusDilemma:
		return "Dilemma"
	case ChannelStatusDilemmaHold:
		return "DilemmaHold"
	case ChannelStatusWatchDog:
		return "WatchDog"
	case ChannelStatusOpenCircuit:
		return "OpenCircuit"
	case ChannelStatusShort:
		return "Short"
	case ChannelStatusFailSafeOff:
		return "FailSafeOff"
	case ChannelStatusFailSafeOn:
		return "FailSafeOn"
	case ChannelStatusForceClear:
		return "ForceClear"
	case ChannelStatusForceSet:
		return "ForceSet"
	default:
		return "Unused"
	}
}
