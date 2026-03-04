package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type indexed struct {
	data map[string]string
}

func (i *indexed) GetValueKey(key string, index string) string {
	radarKey := key + "-" + index
	return radarKey
}

func (i *indexed) logErr(key string, index string, err error, msg string) {
	log.
		Err(err).
		Str("key", i.GetValueKey(key, index)).
		Msg(msg)
}

func (i *indexed) GetAll(key string) []string {
	res := make([]string, 0, 5)

	prefix := key + "-"
	for k, _ := range i.data {
		if strings.HasPrefix(k, prefix) {
			res = append(res, k)
		}
	}
	return res
}

func (i *indexed) HasDefault(key string) bool {
	_, ok := i.data[key]
	return ok
}

func (i *indexed) HasValue(key string, index string) bool {
	fullKey := i.GetValueKey(key, index)
	_, ok := i.data[fullKey]
	return ok
}

func (i *indexed) Has(key string, index string) bool {
	if i.HasValue(key, index) {
		return true
	}

	if i.HasDefault(key) {
		return true
	}

	return false
}

func (i *indexed) SetDefault(key string, value string) {
	i.data[key] = value
}

func (i *indexed) Get(key string, index string, defValue string) string {
	radarKey := i.GetValueKey(key, index)
	if res, ok := i.data[radarKey]; ok {
		return res
	}

	if res, ok := i.data[key]; ok {
		return res
	}

	// Set the default if nothing found
	i.data[key] = defValue

	return defValue
}

func (i *indexed) Set(key string, index string, value string) {
	radarKey := i.GetValueKey(key, index)
	i.data[radarKey] = value
}

func (i *indexed) GetInt(key string, index string, defValue int) int {
	value := i.Get(key, index, strconv.Itoa(defValue))

	res, err := ParseInt(value, defValue)
	if err != nil {
		i.logErr(key, index, err, "unable to parse int")
	}
	return res
}

func (i *indexed) GetFloat(key string, index string, defValue float64) float64 {
	value := i.Get(key, index, strconv.FormatFloat(defValue, 'f', -1, 64))
	res, err := ParseFloat[float64](value, defValue)
	if err != nil {
		i.logErr(key, index, err, "unable to parse float")
	}
	return res
}

func (i *indexed) GetBool(key string, index string, defValue bool) bool {
	var res bool
	var err error

	value := i.Get(key, index, strconv.FormatBool(defValue))

	if res, err = strconv.ParseBool(value); err != nil {
		i.logErr(key, index, err, "unable to parse float")
		return defValue
	}

	return res
}

func (i *indexed) GetDurationMs(key string, index string, defValue int) time.Duration {
	value := i.GetInt(key, index, defValue)
	return time.Duration(value) * time.Millisecond
}
