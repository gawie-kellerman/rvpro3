package port

type StatisticsOutputType uint8

const (
	SotCurrentData StatisticsOutputType = iota
	SotArchiveData
)

func (s StatisticsOutputType) String() string {
	switch s {
	case SotCurrentData:
		return "Current data"
	case SotArchiveData:
		return "Archive data"
	default:
		return "Unknown"
	}
}
