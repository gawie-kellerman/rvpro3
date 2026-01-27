package port

import (
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/bit"
)

type EventTriggerReader struct {
	readerMixin
}

func (r *EventTriggerReader) Init(buffer []byte) {
	r.initBuffer(buffer)
}

func (r *EventTriggerReader) IsSupported() bool {
	switch r.VersionMajor {
	case 4:
		return true
	default:
		return false
	}
}

func (r *EventTriggerReader) GetNofTriggeredRelays() uint8 {
	switch r.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+1)
	default:
		return 0
	}
}

func (r *EventTriggerReader) GetNofTriggeredObjects() uint8 {
	switch r.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+2)
	default:
		return 0
	}
}

func (r *EventTriggerReader) GetFeatureFlags() uint8 {
	switch r.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU8(r.Buffer, r.StartOffset+3)
	default:
		return 0
	}
}

func (r *EventTriggerReader) GetRelays1() uint32 {
	switch r.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU32(r.Buffer, r.Order, r.StartOffset+4)
	default:
		return 0
	}
}

func (r *EventTriggerReader) GetRelays2() uint32 {
	switch r.VersionMajor {
	case 4:
		return utils.OffsetReader.ReadU32(r.Buffer, r.Order, r.StartOffset+8)
	default:
		return 0
	}
}

func (r *EventTriggerReader) PrintDetail() {
	utils.Print.Detail("Event Trigger", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Triggered Objects", "%d\n", r.GetNofTriggeredObjects())
	utils.Print.Detail("Triggered Relays", "%d\n", r.GetNofTriggeredRelays())
	utils.Print.Detail("Feature Flags", "%d\n", r.GetFeatureFlags())
	utils.Print.Detail("Relays 1", "% 10d, %32b\n", r.GetRelays1(), r.GetRelays1())
	utils.Print.Detail("Relays 2", "% 10d, %32b\n", r.GetRelays2(), r.GetRelays2())
	utils.Print.Indent(-2)
}

func (r *EventTriggerReader) TotalSize() int {
	return r.StartOffset + 12
}

func (r *EventTriggerReader) GetRelays() uint64 {
	lo := r.GetRelays1()
	hi := r.GetRelays2()
	return bit.CombineU32(hi, lo)
}
