package port

type StatisticsFeatures uint8

const (
	SfVolume StatisticsFeatures = 1 << iota
	SfOccupancy
	SfAvgSpeed
	SfPercSpeed85th
	SfHeadway
	SfGap
)

func (s StatisticsFeatures) String() string {
	switch s {
	case SfVolume:
		return "Volume"
	case SfOccupancy:
		return "Occupancy"
	case SfAvgSpeed:
		return "AvgSpeed"
	case SfPercSpeed85th:
		return "PercSpeed85th"
	case SfHeadway:
		return "Headway"
	case SfGap:
		return "Gap"
	default:
		return "Unknown"
	}
}
