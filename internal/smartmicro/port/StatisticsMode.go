package port

type StatisticsMode uint8

const (
	SmVolume StatisticsMode = iota
	SmAvgSpeed
	SmPercSpeed85th
	SmOccupancy
	SmHeadway
	SmGap
	SmUnset
)

func (s StatisticsMode) String() string {
	switch s {
	case SmVolume:
		return "Volume"
	case SmOccupancy:
		return "Occupancy"
	case SmAvgSpeed:
		return "AvgSpeed"
	case SmPercSpeed85th:
		return "PercSpeed85th"
	case SmHeadway:
		return "Headway"
	case SmGap:
		return "Gap"
	case SmUnset:
		return "Unset"
	default:
		return "Unknown"
	}
}
