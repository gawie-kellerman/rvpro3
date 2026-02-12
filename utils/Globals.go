package utils

var GlobalSettings Settings
var GlobalMetrics globalMetrics

func init() {
	GlobalSettings.Init()
	GlobalMetrics.Init()
}
