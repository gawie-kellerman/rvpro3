package config

import (
	"io"
	"os"
	"path"

	"rvpro3/radarvision.com/internal/config/globalkey"
	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	jack "gopkg.in/natefinch/lumberjack.v2"
)

type RVProSetup struct{}

func (c *RVProSetup) UDPKeepAlive(o *service.UDPKeepAlive) {
	o.LocalIPAddr = RVPro.GlobalIP(globalkey.UDPKeepAliveCallbackIP)
	o.MulticastIPAddr = RVPro.GlobalIP(globalkey.UDPKeepAliveCastIP)
	o.CooldownMs = RVPro.GlobalInt(globalkey.UDPKeepAliveCooldown)
	o.SendTimeout = RVPro.GlobalInt(globalkey.UDPKeepAliveSendTimeout)
	o.ClientId = uint32(RVPro.GlobalInt(globalkey.UDPKeepAliveClientID))
	o.ReconnectOnCycle = RVPro.GlobalInt(globalkey.UDPKeepAliveReconnectCycle)
	o.LogRepeatMillis = int64(RVPro.GlobalInt(globalkey.UDPKeepAliveLogRepeatMillis))
}

func (c *RVProSetup) UDPData(o *service.UDPData) {
	o.ListenAddr = RVPro.GlobalIP(globalkey.UDPKeepAliveCallbackIP)
	o.ReadTimeout = RVPro.GlobalInt(globalkey.UDPDataReadTimeout)
	o.ReconnectCycle = RVPro.GlobalInt(globalkey.UDPDataReconnectCycle)
	o.ReconnectSleep = RVPro.GlobalInt(globalkey.UDPDataReconnectSleep)
	o.LogRepeatMillis = int64(RVPro.GlobalInt(globalkey.UDPDataLogRepeatMillis))
}

func (c *RVProSetup) Channels(
	o *udp.RadarChannels,
	udpData *service.UDPData,
	// workflows interfaces.IUDPWorkflowBuilder,
) {
	//now := time.Now()
	radarIPsStrings := RVPro.GlobalStrings(globalkey.UDPSupportedRadars, ",")
	numberOfRadars := len(radarIPsStrings)

	instrumentation.GlobalRadarMetrics.Init(numberOfRadars)
	o.Init(numberOfRadars)
	o.AttachTo(udpData)

	for index, radarIPStr := range radarIPsStrings {
		ip := utils.IP4Builder.FromString(radarIPStr)

		//radar := &o.Radar[index]
		o.Radar[index].IPAddress = ip
		//radar.DiagnosticsWorkflow = workflows.GetDiagnosticsWorkflow(radar)
		//radar.InstructionWorkflow = workflows.GetInstructionWorkflow(radar)
		//radar.StatisticsWorkflow = workflows.GetStatisticsWorkflow(radar)
		//radar.TriggersWorkflow = workflows.GetTriggerWorkflow(radar)
		//radar.ObjectListWorkflow = workflows.GetObjectListWorkflow(radar)
		//radar.PvrWorkflow = workflows.GetPVRWorkflow(radar)

		//instrumentation.GlobalRadarMetrics.ByIndex(index).GetRel(int(instrumentation.RmtRadarPort)).SetU32(ip.ToU32(), now)
		//instrumentation.GlobalRadarMetrics.ByIndex(index).GetRel(int(instrumentation.RmtRadarPort)).SetU16(uint16(ip.Port), now)
	}
}

func (c *RVProSetup) ZeroLog() {
	var writers []io.Writer
	zerolog.TimeFieldFormat = utils.DisplayDTMS

	if RVPro.GlobalBool(globalkey.LogToConsole) {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: utils.DisplayDTMS,
		})
	}

	logFileDir := RVPro.GlobalStr(globalkey.LogFileDir)
	logFileName := RVPro.GlobalStr(globalkey.LogFileName)

	if logFileDir != "" && logFileName != "" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        c.rollingAppender(logFileDir, logFileName),
			TimeFormat: utils.DisplayDTMS,
		})
	}

	mw := io.MultiWriter(writers...)
	logger := zerolog.New(mw).With().Timestamp().Logger()

	log.Logger = logger
}

func (c *RVProSetup) rollingAppender(dir string, fn string) io.Writer {
	if err := os.MkdirAll(dir, 0744); err != nil {
		log.Error().
			Err(err).
			Str("logFileDir", dir).
			Msg("Failed to create log directory")
		return nil
	}

	return &jack.Logger{
		Filename:   path.Join(dir, fn),
		MaxSize:    RVPro.GlobalInt(globalkey.LogFileMaxSizeMB),
		MaxAge:     RVPro.GlobalInt(globalkey.LogFileMaxAgeDays),
		MaxBackups: RVPro.GlobalInt(globalkey.LogFileMaxBackups),
	}
}
