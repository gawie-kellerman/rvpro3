package port

type StatisticsOutputType uint8

const (
	SotCurrentData StatisticsOutputType = iota
	SotArchiveData
)

func (s StatisticsOutputType) String() string {
	switch s {
	case SotCurrentData:
		return "Current Data"
	case SotArchiveData:
		return "Archive Data"
	default:
		return "Unknown"
	}
}
