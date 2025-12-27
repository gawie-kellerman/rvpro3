package utils

import (
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
)

const defaultDot = "Default."
const globalDot = "Global."

type KvPairConfigProvider struct {
	data        map[string]string
	indexedName string
}

func (r *KvPairConfigProvider) Init() {
	r.data = make(map[string]string, 100)
}

func (r *KvPairConfigProvider) SetDefault(key string, value string) {
	fullKey := defaultDot + key
	r.data[fullKey] = value
}

func (r *KvPairConfigProvider) SetGlobal(key string, value string) {
	fullKey := globalDot + key
	r.data[fullKey] = value
}

func (r *KvPairConfigProvider) Merge(source map[string]string) {
	for key, value := range source {
		r.data[key] = value
	}
}

func (r *KvPairConfigProvider) GlobalStr(key string) string {
	globalKey := r.getGlobalKey(key)
	if res, ok := r.data[globalKey]; ok {
		return res
	}
	panic(fmt.Sprintf("data not found for key %s.", key))
}

func (r *KvPairConfigProvider) IndexedStr(name string, radar int, key string) string {
	radarKey := r.getIndexedKey(name, radar, key)
	if res, ok := r.data[radarKey]; ok {
		return res
	}

	defaultKey := r.getDefaultKey(key)
	if res, ok := r.data[defaultKey]; ok {
		return res
	}

	panic(fmt.Sprintf("data not found for radar %d, key %s.", radar, key))
}

func (r *KvPairConfigProvider) IndexedInt(name string, radar int, key string) int {
	value := r.IndexedStr(name, radar, key)

	res, err := ParseInt[int](value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(name, radar, key), err)
	}
	return res
}

func (r *KvPairConfigProvider) IndexedFloat(name string, radar int, key string) float64 {
	value := r.IndexedStr(name, radar, key)
	res, err := ParseFloat[float64](value, 0)
	if err != nil {
		r.handleErr(r.getIndexedKey(name, radar, key), err)
	}
	return res
}

func (r *KvPairConfigProvider) IndexedBool(name string, radar int, key string) bool {
	var res bool
	var err error

	value := r.IndexedStr(name, radar, key)

	if res, err = strconv.ParseBool(value); err != nil {
		return false
	}

	return res
}

func (r *KvPairConfigProvider) getIndexedKey(name string, radar int, key string) string {
	radarKey := name + strconv.Itoa(radar) + "." + key
	return radarKey
}

func (r *KvPairConfigProvider) getDefaultKey(key string) string {
	defaultKey := defaultDot + key
	return defaultKey
}

func (r *KvPairConfigProvider) getGlobalKey(key string) string {
	defaultKey := globalDot + key
	return defaultKey
}

func (r *KvPairConfigProvider) SaveToFile(filename string) error {
	return MapSerializer.Save(r.data, filename)
}

func (r *KvPairConfigProvider) MergeFromFile(filename string) error {
	source, err := MapSerializer.Load(filename)
	if err != nil {
		return err
	}

	r.Merge(source)
	return nil
}

func (r *KvPairConfigProvider) DumpTo(stdout *os.File) {
	keys := maps.Keys(r.data)

	res := slices.SortedFunc(keys, func(a string, b string) int {
		return strings.Compare(a, b)
	})

	for _, key := range res {
		_, _ = fmt.Fprintf(stdout, "%s = %s\n", key, r.data[key])
	}
}

func (r *KvPairConfigProvider) GlobalIP(key string) IP4 {
	value := r.GlobalStr(key)

	return IP4Builder.FromString(value)
}

func (r *KvPairConfigProvider) GlobalInt(key string) int {
	value := r.GlobalStr(key)

	res, err := ParseInt[int](value, 0)
	r.handleErr(r.getGlobalKey(key), err)
	return res
}

func (r *KvPairConfigProvider) handleErr(key string, err error) {
	if err != nil {
		log.Fatalf("error %s while parsing key %s", err, key)
	}
}

func (r *KvPairConfigProvider) GlobalStrings(radars string, s string) []string {
	value := r.GlobalStr(radars)

	return strings.Split(value, s)
}

func (r *KvPairConfigProvider) GlobalBool(key string) bool {
	value := r.GlobalStr(key)

	res, err := strconv.ParseBool(value)
	if err != nil {
		r.handleErr(r.getGlobalKey("key"), err)
	}
	return res
}

func (r *KvPairConfigProvider) Set(name string, value string) {
	r.data[name] = value
}
