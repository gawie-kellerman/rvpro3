package port

import (
	"strings"

	"rvpro3/radarvision.com/utils"
)

type FlagsType uint32

const (
	FlMessageCount FlagsType = 1 << iota
	FlTimestamp
	FlSkipPayloadCrc
	FlSourceClientId
	FlTargetClientId
	FlDataIdentifier
	FlSegmentation
)

const sizeOfMessageCount = 2
const sizeOfTimestamp = 8
const sizeOfSourceClientId = 4
const sizeOfTargetClientId = 4
const sizeOfDataIdentifier = 2
const sizeOfSegmentation = 2

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

func (f FlagsType) IsDataIdentifier() bool {
	return f&FlDataIdentifier != 0
}

func (f FlagsType) IsSegmentation() bool {
	return f&FlSegmentation != 0
}

func (f FlagsType) Set(flag FlagsType) FlagsType {
	return f | flag
}

func (f FlagsType) String() string {
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

	if f.IsDataIdentifier() {
		result.WriteString("data identifier,")
	}

	if f.IsSegmentation() {
		result.WriteString("segmentation,")
	}

	return result.String()
}

func (f FlagsType) SizeOf() uint8 {
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

	if f.IsDataIdentifier() {
		result += 2
	}

	if f.IsSegmentation(nil) {
		result += 2
	}

	return result
}

func (f FlagsType) PrintDetail(th *TransportHeader) {
	if f.IsMessageCount() {
		utils.Print.Detail("Message Counter", "%d\n", th.MessageCounter)
	}
	if f.IsTimestamp() {
		utils.Print.Detail("Timestamp", "%d\n", th.Timestamp)
	}
	if f.IsSourceClientId() {
		utils.Print.Detail("Source Client Id", "0x%x\n", th.SourceClientId)
	}
	if f.IsTargetClientId() {
		utils.Print.Detail("Target Client Id", "0x%x\n", th.TargetClientId)
	}
}

func (f FlagsType) OffsetOf(upTo FlagsType) int {
	res := 0

	if upTo.IsMessageCount() {
		return res
	}

	if f.IsMessageCount() {
		res += sizeOfMessageCount
	}

	if upTo.IsTimestamp() {
		return res
	}

	if f.IsTimestamp() {
		res += sizeOfTimestamp
	}

	if upTo.IsSourceClientId() {
		return res
	}

	if f.IsSourceClientId() {
		res += sizeOfSourceClientId
	}

	if upTo.IsTargetClientId() {
		return res
	}

	if f.IsTargetClientId() {
		res += sizeOfTargetClientId
	}

	if upTo.IsDataIdentifier() {
		return res
	}

	if f.IsDataIdentifier() {
		res += sizeOfDataIdentifier
	}

	if upTo.IsSegmentation() {
		return res
	}

	panic("unreachable")
}
