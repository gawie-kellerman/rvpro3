package trigger

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/udp/state"
	"rvpro3/radarvision.com/utils"
)

const rawCSVEnabled = "activity.trigger.unmasked.csv.enabled"
const rawCSVPathTemplate = "activity.trigger.unmasked.csv.pathtemplate"
const rawCSVPathDefault = "/media/SDLOGS/logs/sensor/%d/events/events-%%s.csv"

type LogCSVActivity struct {
	interfaces.UDPActivityMixin
	Metrics    LogCSVActivityMetrics
	CSVWriter  TriggerCSVWriter
	CSVError   utils.ErrorLoggerMixin
	radarState state.RadarState
	IsEnabled  bool
	OldRelays  uint64
}

type LogCSVActivityMetrics struct {
	ErrorMajorMinorVersion *utils.Metric
	ProcessCount           *utils.Metric
	SkipEqualsCount        *utils.Metric
	SkipDisabledCount      *utils.Metric
	utils.MetricsInitMixin
}

func (l *LogCSVActivity) Init(workflow interfaces.IUDPWorkflow, index int, fullName string) {
	l.InitBase(workflow, index, fullName)
	l.Metrics.InitMetrics(fullName, &l.Metrics)

	radarIP := l.Workflow.GetRadarIP()
	radarState := state.RadarStateHelper.GetOrSet(radarIP)
	if radarState != nil {
		l.CSVWriter.SensorName = radarState.Name
	}
	gs := &utils.GlobalSettings

	// Load IsEnabled
	l.IsEnabled = gs.Indexed.GetBool(rawCSVEnabled, radarIP.String(), true)

	// Setup CSVWriter
	l.CSVWriter.SensorIP = radarIP.String()
	l.CSVWriter.CSVFacade.PathTemplate = gs.Indexed.Get(
		rawCSVPathTemplate,
		radarIP.String(),
		fmt.Sprintf(rawCSVPathDefault, radarIP.GetHost()),
	)
	l.CSVWriter.Init()
	l.OldRelays = 1
}

func (l *LogCSVActivity) Process(time time.Time, bytes []byte) {
	if l.IsEnabled {
		th := port.TransportHeaderReader{
			Buffer: bytes,
		}

		ph := port.PortHeaderReader{
			Buffer:      bytes,
			StartOffset: int(th.GetHeaderLength()),
		}

		if ph.GetPortMajorVersion() == 4 && ph.GetPortMinorVersion() == 0 {
			trigger := port.EventTriggerReader{}
			trigger.Init(bytes)

			relays := trigger.GetRelays()

			if relays == l.OldRelays {
				l.Metrics.SkipEqualsCount.IncAt(1, time)
				return
			}

			l.CSVWriter.SensorSerial = l.radarState.ReplaceSerial(th.GetSourceClientId())

			if err := l.CSVWriter.Write(
				time,
				int(trigger.GetNofTriggeredObjects()),
				int(trigger.GetNofTriggeredRelays()),
				relays,
			); err != nil {
				msg := fmt.Sprintf(
					"trigger for %s log to csv failed: %s",
					l.Workflow.GetRadarIP(),
					err.Error(),
				)
				l.CSVError.LogErrorAt(time, msg, err)
			}

			l.OldRelays = relays
			l.Metrics.ProcessCount.Inc(1)
		} else {
			l.Metrics.ErrorMajorMinorVersion.IncAt(1, time)
		}
	} else {
		l.Metrics.SkipDisabledCount.IncAt(1, time)
	}
}
