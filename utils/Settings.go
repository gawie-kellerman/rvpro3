package utils

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

// Settings is a highly opinionated with many side effects.  All methods, in
// the absence of the key found, will inject the 'new default' value into the map.  To avoid
// map updates, use the Has-methods before using Get-methods.
type Settings struct {
	data map[string]string
	//indexedName string

	Indexed indexed
	Basic   basic
}

func (r *Settings) Init() {
	r.data = make(map[string]string, 100)
	r.Indexed.data = r.data
	r.Basic.data = r.data
}

func (r *Settings) Merge(source map[string]string) {
	maps.Copy(r.data, source)
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

func (r *Settings) MergeFromSettings(settings *Settings) {
	r.Merge(settings.data)
}

func (r *Settings) ReadArgs() {
	overrides := Args.GetKVPairIndexes("--override|-o")

	for _, override := range overrides {
		key := Args.GetKeyName(override, "--override|-o")
		_, value := Args.GetPair("--override|-o", key)
		if key != "" {
			r.data[key] = value
		}
	}
}

func (r *Settings) Split(value string) []string {
	return strings.Split(value, ";")
}
