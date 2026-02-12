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

type Settings struct {
	data        map[string]string
	indexedName string
}

func (r *Settings) Init() {
	r.data = make(map[string]string, 100)
}

func (r *Settings) SetDefault(key string, value string) {
	fullKey := defaultDot + key
	r.data[fullKey] = value
}

func (r *Settings) SetSetting(key string, value string) {
	fullKey := settingDot + key
	r.data[fullKey] = value
}

func (r *Settings) Merge(source map[string]string) {
	maps.Copy(r.data, source)
}

func (r *Settings) GetSettingAsStr(key string) string {
	globalKey := r.getSettingKey(key)
	if res, ok := r.data[globalKey]; ok {
		return res
	}
	panic(fmt.Sprintf("data not found for key %s.", key))
}

func (r *Settings) GetIndexedAsStr(entityName string, entityIndex string, configKey string) string {
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

func (r *Settings) GetIndexedAsInt(entityName string, entityIndex string, configKey string) int {
	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)

	res, err := ParseInt(value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(entityName, entityIndex, configKey), err)
	}
	return res
}

func (r *Settings) GetIndexedAsFloat(entityName string, entityIndex string, configKey string) float64 {
	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)
	res, err := ParseFloat[float64](value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(entityName, entityIndex, configKey), err)
	}
	return res
}

func (r *Settings) GetIndexedAsBool(entityName string, entityIndex string, configKey string) bool {
	var res bool
	var err error

	value := r.GetIndexedAsStr(entityName, entityIndex, configKey)

	if res, err = strconv.ParseBool(value); err != nil {
		return false
	}

	return res
}

func (r *Settings) getIndexedKey(entityName string, entityIndex string, configKey string) string {
	radarKey := entityName + "." + entityIndex + "." + configKey
	return radarKey
}

func (r *Settings) getDefaultKey(key string) string {
	defaultKey := defaultDot + key
	return defaultKey
}

func (r *Settings) getSettingKey(key string) string {
	defaultKey := settingDot + key
	return defaultKey
}

func (r *Settings) SaveToFile(filename string) error {
	return MapSerializer.Save(r.data, filename)
}

func (r *Settings) MergeFromFile(filename string) error {
	source, err := MapSerializer.Load(filename)
	if err != nil {
		return err
	}

	r.Merge(source)
	return nil
}

func (r *Settings) DumpTo(stdout *os.File) {
	keys := maps.Keys(r.data)

	res := slices.SortedFunc(keys, func(a string, b string) int {
		return strings.Compare(a, b)
	})

	for _, key := range res {
		_, _ = fmt.Fprintf(stdout, "%s = %s\n", key, r.data[key])
	}
}

func (r *Settings) GetSettingAsIP(key string) IP4 {
	value := r.GetSettingAsStr(key)

	return IP4Builder.FromString(value)
}

func (r *Settings) handleErr(key string, err error) {
	if err != nil {
		log.Fatalf("error %s while parsing key %s", err, key)
	}
}

func (r *Settings) GetSettingAsSplit(key string, delimiter string) []string {
	value := r.GetSettingAsStr(key)

	return strings.Split(value, delimiter)
}

func (r *Settings) SetRaw(name string, value string) {
	r.data[name] = value
}

func (r *Settings) SetSettingAsBool(key string, value bool) {
	strValue := strconv.FormatBool(value)
	r.SetSetting(key, strValue)
}

func (r *Settings) GetSettingAsBool(key string) bool {
	value := r.GetSettingAsStr(key)

	res, err := strconv.ParseBool(value)
	if err != nil {
		r.handleErr(r.getSettingKey("key"), err)
	}
	return res
}

func (r *Settings) GetOrPutBool(key string, defValue bool) bool {
	globalKey := r.getSettingKey(key)

	setting, ok := r.data[globalKey]
	if !ok {
		r.data[globalKey] = strconv.FormatBool(defValue)
		return defValue
	}

	res, err := strconv.ParseBool(setting)
	if err != nil {
		r.handleErr(r.getSettingKey("key"), err)
	}
	return res
}

func (r *Settings) GetOrPutStr(key string, defValue string) string {
	globalKey := r.getSettingKey(key)

	value, ok := r.data[globalKey]
	if !ok {
		r.data[globalKey] = defValue
		return defValue
	}

	return value
}

func (r *Settings) SetSettingAsStr(key string, value string) {
	r.SetSetting(key, value)
}

func (r *Settings) GetSettingAsInt(key string) int {
	value := r.GetSettingAsStr(key)

	res, err := ParseInt(value, 0)
	r.handleErr(r.getSettingKey(key), err)
	return res
}

func (r *Settings) SetSettingAsInt(key string, value int) {
	strValue := strconv.Itoa(value)
	r.SetSetting(key, strValue)
}

func (r *Settings) SetSettingAsMillis(key string, milliseconds int) {
	strValue := strconv.Itoa(milliseconds)
	r.SetSetting(key, strValue)
}

func (r *Settings) GetSettingAsMillis(key string) time.Duration {
	value := r.GetSettingAsInt(key)
	return time.Duration(value) * time.Millisecond
}

func (r *Settings) MergeFromSettings(settings *Settings) {
	r.Merge(settings.data)
}
