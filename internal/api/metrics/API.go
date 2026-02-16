package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"rvpro3/radarvision.com/utils"
)

type API struct {
	BaseURL string
}

func NewMetricsAPI(baseURL string) *API {
	return &API{BaseURL: baseURL}
}

func (r *API) GetUDPBroker(radarIP utils.IP4) (res *UDPBroker, err error) {
	url := fmt.Sprintf("%s/metrics/section?sn=UDP.Broker.%s", r.BaseURL, radarIP)
	var response *http.Response

	client := http.DefaultClient
	response, err = client.Get(url)
	utils.Debug.Panic(err)
	defer response.Body.Close()

	buffer, err := io.ReadAll(response.Body)
	utils.Debug.Panic(err)

	res = &UDPBroker{}
	err = json.Unmarshal(buffer, &res.Metrics)
	return res, err
}
