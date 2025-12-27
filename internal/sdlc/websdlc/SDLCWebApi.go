package websdlc

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"rvpro3/radarvision.com/utils/bit"
)

const MetricsAt = "SDLC.WEB"

type sdlcWebApi struct{}

var SDLCWebApi sdlcWebApi

func (sdlcWebApi) GetStatus4(basePath string) (res Status4Response, err error) {
	var fullUrl string
	var httpResponse *http.Response
	var httpBody []byte

	fullUrl, err = url.JoinPath(basePath, "/status4.xml")

	if err != nil {
		return res, err
	}

	if httpResponse, err = http.Get(fullUrl); err != nil {
		return res, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return res, errors.New(httpResponse.Status)
	}

	if httpBody, err = io.ReadAll(httpResponse.Body); err != nil {
		return res, err
	}

	if err = xml.Unmarshal(httpBody, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (sdlcWebApi) SendTS2Detect(basePath string, trigger uint64, mask uint64) (res SendDetectResponse, err error) {
	var httpBody []byte
	var httpResponse *http.Response
	var uri *url.URL
	detect := SDLCWebApi.GetTS2Detect(trigger, mask)

	fullUrl, err := url.JoinPath(basePath, "/dets.cgi")

	if err != nil {
		return res, err
	}
	uri, err = url.Parse(fullUrl)

	if err != nil {
		return res, err
	}

	params := uri.Query()
	params.Add("det", detect)
	uri.RawQuery = params.Encode()

	if httpResponse, err = http.Get(uri.String()); err != nil {
		return res, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return res, errors.New(httpResponse.Status)
	}

	if httpBody, err = io.ReadAll(httpResponse.Body); err != nil {
		return res, err
	}

	err = res.Unmarshal(string(httpBody))
	return res, err
}

func (sdlcWebApi) GetTS2Detect(trigger uint64, mask uint64) string {
	var buffer [64]byte

	bit.ForLSB(mask, func(index int, on bool) {
		if on {
			if bit.IsSet(trigger, index) {
				buffer[index] = '1'
			} else {
				buffer[index] = '0'
			}
		} else {
			buffer[index] = '2'
		}
	})

	return string(buffer[:])
}
