package utils

import (
	"reflect"
	"strings"
)

type MetricsInitMixin struct {
}

func (m *MetricsInitMixin) InitMetrics(sectionName string, source any) {
	gm := &GlobalMetrics

	elem := reflect.ValueOf(source).Elem()
	metricType := reflect.TypeOf(&Metric{})

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Type() == metricType {
			fieldName := elem.Type().Field(i).Name

			var metric *Metric

			if strings.Contains(fieldName, "Time") {
				metric = gm.Metric(sectionName, fieldName, MtMilliTime)
			} else if strings.Contains(fieldName, "Dur") {
				metric = gm.Metric(sectionName, fieldName, MtMilliDuration)
			} else if strings.HasPrefix(fieldName, "Is") {
				metric = gm.Metric(sectionName, fieldName, MtBool)
			} else {
				metric = gm.Metric(sectionName, fieldName, MtI64)
			}

			field.Set(reflect.ValueOf(metric))
		}
	}
}
