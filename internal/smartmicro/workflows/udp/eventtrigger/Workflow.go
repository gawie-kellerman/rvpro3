package eventtrigger

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
)

type Workflow struct {
	interfaces.MixinWorkflow
}

func (w *Workflow) Process(time time.Time, bytes []byte) {
	reader := port.EventTriggerReader{}
	reader.Init(bytes)

	channel := w.GetParent().(*udp.RadarChannel)
	state := &channel.State

	if state.Trigger.Update(time, reader.GetRelays1(), reader.GetRelays2()) {
		// TODO: Update the data metric...
		//metric := channel.Metrics.GetRel(int(instrumentation.RmtDiagnosticProcessed))
		//channel.Metrics.SetU32s(reader.GetRelays1(), reader.GetRelays2())
	}

	// Down-line:
	// 1. Can save the triggers upon update after receive
	// 2. Process the trigger for potential red hold
	// 3. Can save the triggers after processing
}
