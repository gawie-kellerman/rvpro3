package globalkey

const UDPDataConnectionCooldown = "UDPData.Connection.Cooldown"
const UDPDataConnectionCooldownDefault = 3000
const StartupMode = "Startup.Mode"
const StartupModeUDP = "UDP"
const StartupModeDefault = StartupModeUDP

const UDPKeepAliveCallbackIP = "UDP.KeepAlive.IP"
const UDPKeepAliveCastIP = "UDP.KeepAlive.CastIP"
const UDPKeepAliveCooldown = "UDP.KeepAlive.Cooldown"
const UDPKeepAliveSendTimeout = "UDP.KeepAlive.Timeout"
const UDPKeepAliveReconnectCycle = "UDP.KeepAlive.ReconnectCycle"
const UDPKeepAliveClientID = "UDP.KeepAlive.ClientID"
const UDPKeepAliveLogRepeatMillis = "UDP.KeepAlive.RepeatMillis"
const UDPKeepAliveLogRepeatMillisDefault = "60000"

const UDPDataReadTimeout = "UDP.Data.ReadTimeout"
const UDPDataReconnectCycle = "UDP.Data.Reconnect.Cycle"
const UDPDataReconnectSleep = "UDP.Data.Reconnect.Sleep"
const UDPDataLogRepeatMillis = "UDP.Data.Log.RepeatMillis"
const UDPDataLogRepeatMillisDefault = "60000"

const UDPSupportedRadars = "UDP.SupportedRadars"

const UDPKeepAliveCallbackIPDefault = "192.168.11.2:55555"
const UDPKeepAliveCastIPDefault = "239.144.0.0:60000"
const UDPKeepAliveCooldownDefault = "1000"
const UDPKeepAliveSendTimeoutDefault = "1000"
const UDPKeepAliveReconnectCycleDefault = "5"
const UDPKeepAliveClientIDDefault = "0x1000001"

const UDPDataReadTimeoutDefault = "3000"
const UDPDataReconnectSleepDefault = "1000"

const UDPSupportedRadarsDefault = "192.168.11.12:55555,192.168.11.13:55555,192.168.11.14:55555,192.168.11.15:55555"
const UDPDataReconnectCycleDefault = "3"

const HttpHost = "Http.Host"
const HttpHostDefault = "localhost:8080"

const LogLevel = "Log.Level"
const LogLevelDefault = "info"
const LogToConsole = "Log.To.Console"
const LogToConsoleDefault = "false"
const LogFileDir = "Log.File.Dir"
const LogFileDirDefault = "/media/SDLOGS/logs/system"
const LogFileName = "Log.File.Name"
const LogFileNameDefault = "rvm.log"
const LogFileMaxSizeMB = "Log.File.MaxSizeMB"
const LogFileMaxSizeMBDefault = "10"
const LogFileMaxBackups = "Log.File.MaxBackups"
const LogFileMaxBackupsDefault = "10"
const LogFileMaxAgeDays = "Log.File.MaxAgeDays"
const LogFileMaxAgeDaysDefault = "30"

const IsLCDEnabled = "Is.LCD.Enabled"
const IsSDLCEnabled = "Is.SDLC.Enabled"
const IsSNMPEnabled = "Is.SNMP.Enabled"
const IsHubEnabled = "Is.Hub.Enabled"
