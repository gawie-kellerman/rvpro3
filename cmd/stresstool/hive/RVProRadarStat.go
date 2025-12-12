package hive

import "math"

type RVProRadarStat struct {
	RadarIP         string
	DataCount       uint64
	ObjectLists     uint64
	EventTriggers   uint64
	StatisticsCount uint64
	PVRCount        uint64
}

func (r *RVProRadarStat) Parse(node map[string]any) {
	readInt := func(key string) uint64 {
		return uint64(math.Round(node[key].(float64)))
	}
	r.RadarIP = node["ip_addr"].(string)
	r.DataCount = readInt("data_cnt")
	r.ObjectLists = readInt("object_lists_count")
	r.EventTriggers = readInt("event_triggers_count")
	r.StatisticsCount = readInt("statistics_count")
}
