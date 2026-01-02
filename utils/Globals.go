package utils

var GlobalConfig Config
var GlobalMetrics globalMetrics

func init() {
	GlobalConfig.Init()
	GlobalMetrics.Init()
}
