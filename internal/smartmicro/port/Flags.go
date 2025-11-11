package port

import "strings"

type FlagsType uint32

const (
	FlMessageCount FlagsType = 1 << iota
	FlTimestamp
	FlSkipPayloadCrc
	FlSourceClientId
	FlTargetClientId
)

func (f FlagsType) IsMessageCount() bool {
	return f&FlMessageCount != 0
}

func (f FlagsType) IsTimestamp() bool {
	return f&FlTimestamp != 0
}

func (f FlagsType) IsSkipPayloadCrc() bool {
	return f&FlSkipPayloadCrc != 0
}

func (f FlagsType) IsSourceClientId() bool {
	return f&FlSourceClientId != 0
}

func (f FlagsType) IsTargetClientId() bool {
	return f&FlTargetClientId != 0
}

func (f FlagsType) Set(flag FlagsType) FlagsType {
	return f | flag
}

func (f FlagsType) ToString() string {
	result := strings.Builder{}
	result.Grow(100)

	if f.IsMessageCount() {
		result.WriteString("message count,")
	}
	if f.IsTimestamp() {
		result.WriteString("timestamp,")
	}
	if f.IsSkipPayloadCrc() {
		result.WriteString("skip payload crc,")
	}
	if f.IsSourceClientId() {
		result.WriteString("source clientId,")
	}
	if f.IsTargetClientId() {
		result.WriteString("target clientId,")
	}

	return result.String()
}

func (f FlagsType) SizeOf(flags FlagsType) uint8 {
	result := uint8(0)

	if f.IsMessageCount() {
		result += 2
	}
	if f.IsTimestamp() {
		result += 2
	}
	if f.IsSourceClientId() {
		result += 4
	}
	if f.IsTargetClientId() {
		result += 4
	}

	return result
}
