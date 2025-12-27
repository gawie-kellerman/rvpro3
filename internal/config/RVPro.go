package config

import (
	"strconv"

	"rvpro3/radarvision.com/internal/config/globalkey"
	"rvpro3/radarvision.com/internal/config/radarkey"
	"rvpro3/radarvision.com/utils"
)

var RVPro utils.KvPairConfigProvider

const Radar = "Radar"

func init() {
	RVPro.Init()
	RVPro.SetGlobal(globalkey.StartupMode, globalkey.StartupModeDefault)

	// Supported Radars
	RVPro.SetGlobal(globalkey.UDPSupportedRadars, globalkey.UDPSupportedRadarsDefault)

	// UDP KeepAlive
	RVPro.SetGlobal(globalkey.UDPKeepAliveCallbackIP, globalkey.UDPKeepAliveCallbackIPDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveCastIP, globalkey.UDPKeepAliveCastIPDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveCooldown, globalkey.UDPKeepAliveCooldownDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveSendTimeout, globalkey.UDPKeepAliveSendTimeoutDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveReconnectCycle, globalkey.UDPKeepAliveReconnectCycleDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveClientID, globalkey.UDPKeepAliveClientIDDefault)
	RVPro.SetGlobal(globalkey.UDPKeepAliveLogRepeatMillis, globalkey.UDPKeepAliveLogRepeatMillisDefault)

	// UDP Data
	RVPro.SetGlobal(globalkey.UDPDataConnectionCooldown, strconv.Itoa(globalkey.UDPDataConnectionCooldownDefault))
	RVPro.SetGlobal(globalkey.UDPDataReconnectCycle, globalkey.UDPDataReconnectCycleDefault)
	RVPro.SetGlobal(globalkey.UDPDataReconnectSleep, globalkey.UDPDataReconnectSleepDefault)
	RVPro.SetGlobal(globalkey.UDPDataReadTimeout, globalkey.UDPDataReadTimeoutDefault)
	RVPro.SetGlobal(globalkey.UDPDataLogRepeatMillis, globalkey.UDPDataLogRepeatMillisDefault)

	// HTTP Host
	RVPro.SetGlobal(globalkey.HttpHost, globalkey.HttpHostDefault)

	// Messages
	RVPro.SetDefault(radarkey.TriggerPath, radarkey.TriggerPathDefault)
	RVPro.SetDefault(radarkey.StatisticsPath, radarkey.StatisticsPathDefault)
	RVPro.SetDefault(radarkey.ObjectListPath, radarkey.ObjectListPathDefault)

	// Log
	RVPro.SetGlobal(globalkey.LogLevel, globalkey.LogLevelDefault)
	RVPro.SetGlobal(globalkey.LogToConsole, globalkey.LogToConsoleDefault)
	RVPro.SetGlobal(globalkey.LogFileDir, globalkey.LogFileDirDefault)
	RVPro.SetGlobal(globalkey.LogFileName, globalkey.LogFileNameDefault)
	RVPro.SetGlobal(globalkey.LogFileMaxSizeMB, globalkey.LogFileMaxSizeMBDefault)
	RVPro.SetGlobal(globalkey.LogFileMaxBackups, globalkey.LogFileMaxBackupsDefault)
	RVPro.SetGlobal(globalkey.LogFileMaxAgeDays, globalkey.LogFileMaxAgeDaysDefault)
}
