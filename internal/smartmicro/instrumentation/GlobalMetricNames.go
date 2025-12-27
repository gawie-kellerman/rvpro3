package instrumentation

var globalMetricNames map[int]string

func GlobalMetricNames() map[int]string {
	if globalMetricNames == nil {
		globalMetricNames = make(map[int]string, 200)
	}
	return globalMetricNames
}

func GlobalMetricName(id int) string {
	res, ok := GlobalMetricNames()[id]
	if !ok {
		res = "Unassigned metric name"
	}
	return res
}
