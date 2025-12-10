package port

import "strings"

type StatisticsFailSafe uint8

const (
	SfsRain StatisticsFailSafe = 1 << iota
	SfsInterference
	sfsFiller1
	sfsFiller2
	SfsBlind
)

func (s StatisticsFailSafe) String() string {
	sb := strings.Builder{}
	sb.Grow(50)

	if s&SfsRain == SfsRain {
		sb.WriteString("rain,")
	}

	if s&SfsInterference == SfsInterference {
		sb.WriteString("interference,")
	}
	if s&SfsBlind == SfsBlind {
		sb.WriteString("blind")
	}
	return sb.String()
}
