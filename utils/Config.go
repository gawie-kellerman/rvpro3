package utils

import (
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultDot = "Default."
const settingDot = "Setting."

type Config struct {
	data        map[string]string
	indexedName string
}

func (r *Config) Init() {
	r.data = make(map[string]string, 100)
}

func (r *Config) SetDefault(key string, value string) {
	fullKey := defaultDot + key
	r.data[fullKey] = value
}

func (r *Config) SetSetting(key string, value string) {
	fullKey := settingDot + key
	r.data[fullKey] = value
}

func (r *Config) Merge(source map[string]string) {
	for key, value := range source {
		r.data[key] = value
	}
}

func (r *Config) GetSettingAsStr(key string) string {
	globalKey := r.getSettingKey(key)
	if res, ok := r.data[globalKey]; ok {
		return res
	}
	panic(fmt.Sprintf("data not found for key %s.", key))
}

func (r *Config) GetIndexedAsStr(entityName string, entityIndex string, configKey string) string {
	radarKey := r.getIndexedKey(entityName, entityIndex, configKey)
	if res, ok := r.data[radarKey]; ok {
		return res
	}

	defaultKey := r.getDefaultKey(configKey)
	if res, ok := r.data[defaultKey]; ok {
		return res
	}

	panic(fmt.Sprintf("data not found for entityIndex %s, configKey %s.", entityIndex, configKey))
}

func (r *Config) GetIndexedAsInt(entityName string, entityIndex string, configKey string) int {
	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)

	res, err := ParseInt[int](value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(entityName, entityIndex, configKey), err)
	}
	return res
}

func (r *Config) GetIndexedAsFloat(entityName string, entityIndex string, configKey string) float64 {
	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)
	res, err := ParseFloat[float64](value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(entityName, entityIndex, configKey), err)
	}
	return res
}

func (r *Config) GetIndexedAsBool(entityName string, entityIndex string, configKey string) bool {
	var res bool
	var err error

	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)

	if res, err = strconv.ParseBool(value); err != nil {
		return false
	}

	return res
}

func (r *Config) getIndexedKey(entityName string, entityIndex string, configKey string) string {
	radarKey := entityName + "." + entityIndex + "." + configKey
	return radarKey
}

func (r *Config) getDefaultKey(key string) string {
	defaultKey := defaultDot + key
	return defaultKey
}

func (r *Config) getSettingKey(key string) string {
	defaultKey := settingDot + key
	return defaultKey
}

func (r *Config) SaveToFile(filename string) error {
	return MapSerializer.Save(r.data, filename)
}

func (r *Config) MergeFromFile(filename string) error {
	source, err := MapSerializer.Load(filename)
	if err != nil {
		return err
	}

	r.Merge(source)
	return nil
}

func (r *Config) DumpTo(stdout *os.File) {
	keys := maps.Keys(r.data)

	res := slices.SortedFunc(keys, func(a string, b string) int {
		return strings.Compare(a, b)
	})

	for _, key := range res {
		_, _ = fmt.Fprintf(stdout, "%s = %s\n", key, r.data[key])
	}
}

func (r *Config) GetSettingAsIP(key string) IP4 {
	value := r.GetSettingAsStr(key)

	return IP4Builder.FromString(value)
}

func (r *Config) handleErr(key string, err error) {
	if err != nil {
		log.Fatalf("error %s while parsing key %s", err, key)
	}
}

func (r *Config) GetSettingAsSplit(key string, delimiter string) []string {
	value := r.GetSettingAsStr(key)

	return strings.Split(value, delimiter)
}

func (r *Config) SetRaw(name string, value string) {
	r.data[name] = value
}

func (r *Config) SetSettingAsBool(key string, value bool) {
	strValue := strconv.FormatBool(value)
	r.SetSetting(key, strValue)
}

func (r *Config) GetSettingAsBool(key string) bool {
	value := r.GetSettingAsStr(key)

	res, err := strconv.ParseBool(value)
	if err != nil {
		r.handleErr(r.getSettingKey("key"), err)
	}
	return res
}

func (r *Config) SetSettingAsStr(key string, value string) {
	r.SetSetting(key, value)
}

func (r *Config) GetSettingAsInt(key string) int {
	value := r.GetSettingAsStr(key)

	res, err := ParseInt[int](value, 0)
	r.handleErr(r.getSettingKey(key), err)
	return res
}

func (r *Config) SetSettingAsInt(key string, value int) {
	strValue := strconv.Itoa(value)
	r.SetSetting(key, strValue)
}

func (r *Config) SetSettingAsMillis(key string, milliseconds int) {
	strValue := strconv.Itoa(milliseconds)
	r.SetSetting(key, strValue)
}

func (r *Config) GetSettingAsMillis(key string) time.Duration {
	value := r.GetSettingAsInt(key)
	return time.Duration(value) * time.Millisecond
}
