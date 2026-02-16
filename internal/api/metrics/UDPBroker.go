package metrics

import (
	"encoding/json"
	"fmt"

	"rvpro3/radarvision.com/utils"
)

type UDPBroker struct {
	Metrics map[string]utils.Metrics
}

func (r *UDPBroker) GetMetric(rootName string, metricName string) *utils.Metric {
	rootNode, ok := r.Metrics[rootName]
	if ok {
		res := rootNode.Get(metricName)
		if res != nil {
			res.Name = metricName
		}
		return res
	}
	return nil
}

func (r *UDPBroker) GetBroker(ip4 utils.IP4, metricName string) *utils.Metric {
	rootName := fmt.Sprintf("UDP.Broker.%s", ip4)
	return r.GetMetric(rootName, metricName)
}

func (r *UDPBroker) LoadJson(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, &r.Metrics)
	return err
}
