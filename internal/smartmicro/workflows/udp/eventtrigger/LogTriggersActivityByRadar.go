package eventtrigger

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/branding"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/triggerpipeline"
	"rvpro3/radarvision.com/utils"
)

type LogTriggersActivityByRadar struct {
	csv           utils.CSVRollOverFileWriterProvider
	CSVTemplate   string
	CSVFormat     string
	CSVFileType   string
	CSVVersion    string
	PipelineLen   int
	PipelineItems [20]triggerpipeline.ITriggerPipelineItem
	parent        interfaces.IUDPWorkflowParent
	Pipeline      *triggerpipeline.TriggerPipeline

	PreviousTriggers utils.Uint128
	CurrentTriggers  utils.Uint128
}

func (l *LogTriggersActivityByRadar) Init(parent interfaces.IUDPWorkflowParent) {
	l.parent = parent
	l.csv.Template = l.CSVTemplate
	l.csv.Format = l.CSVFormat
	l.csv.OnHeader = l.OnCSVHeader
}

func (l *LogTriggersActivityByRadar) Process(now time.Time, readerAny any) {
	reader, ok := readerAny.(*port.EventTriggerReader)

	if !ok {
		panic(errors.New("reader is not an EventTriggerReader"))
	}

	if l.PipelineLen == 0 {
		l.Pipeline = utils.GlobalState.Get(triggerpipeline.TriggerPipelineStateName).(*triggerpipeline.TriggerPipeline)
		l.PipelineLen = l.Pipeline.ListByRadar(l.parent.GetRadarIP(), l.PipelineItems[:])
	}

	l.PreviousTriggers = l.CurrentTriggers
	l.CurrentTriggers = l.Pipeline.ExecuteList(l.PipelineItems[:l.PipelineLen])

	if !l.PreviousTriggers.Equals(l.CurrentTriggers) {
		l.logTriggers(reader)
	}
}

func (l *LogTriggersActivityByRadar) OnCSVHeader(
	provider *utils.CSVRollOverFileWriterProvider,
	writer *utils.CSVWriter,
	_ string,
	_ string,
) {
	writer, err := provider.GetWriter()
	if err != nil {
		log.Err(err).Msgf("Error creating CSV roll over file")
	}

	branding.CSVBranding.WriteTitle(writer, l.CSVFileType, l.CSVVersion)
	branding.CSVBranding.WriteFeaturesNL(writer)
	writer.WriteColsNL("TIMESTAMP", "OBJECTS", "RELAYS", "RELAYS 1", "RELAYS 2")
}

func (l *LogTriggersActivityByRadar) logTriggers(reader *port.EventTriggerReader) {
	writer, err := l.csv.GetWriter()
	if err != nil {
		log.Err(err).Msgf("Error creating CSV roll over file")
	}

	writer.WriteCol(time.Now().Format(utils.DisplayDateTimeMS))
	writer.WriteInt(int(reader.GetNofTriggeredObjects()))
	writer.WriteInt(int(reader.GetNofTriggeredRelays()))
	writer.WriteInt(int(reader.GetRelays1()))
	writer.WriteInt(int(reader.GetRelays2()))
	writer.WriteNL()
	_ = writer.Flush()
}
