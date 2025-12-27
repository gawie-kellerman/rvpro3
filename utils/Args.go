package utils

import (
	"os"
	"strings"
)

type args struct {
}

var Args args

func (args) GetString(names string, defValue string) string {
	index, name := Args.IndexOf(names)

	if index == -1 {
		return defValue
	}

	arg := os.Args[index]

	//--file vs --file= vs --file=whatever
	if len(name)+1 >= len(arg) {
		return defValue
	}

	return arg[len(name)+1:]
}

func (args) Has(names string) bool {
	index, _ := Args.IndexOf(names)
	return index != -1
}

func (args) SplitNames(names string) []string {
	return strings.Split(names, "|")
}

func (args) IndexOf(pipedNames string) (int, string) {
	names := Args.SplitNames(pipedNames)
	for _, name := range names {
		argIndex := Args.IndexOfName(name)
		if argIndex != -1 {
			return argIndex, name
		}
	}

	return -1, ""
}

func (args) IndexOfName(name string) int {
	for index, value := range os.Args[1:] {
		if strings.HasPrefix(value, name) {
			return index + 1
		}
	}

	return -1
}

func (a args) GetKVPairIndexes(pipedNames string) (res []int) {
	names := Args.SplitNames(pipedNames)

	for index, value := range os.Args[1:] {
		for _, name := range names {
			if strings.HasPrefix(value, name) {
				res = append(res, index+1)
				break
			}
		}
	}
	return res
}

func (a args) GetKeyName(index int, pipedNames string) string {
	source := os.Args[index]
	names := Args.SplitNames(pipedNames)

	for _, name := range names {
		if strings.HasPrefix(source, name) {
			interim := source[len(name):]
			eqIndex := strings.Index(interim, "=")
			if eqIndex != -1 {
				return interim[:eqIndex]
			}
			return interim
		}
	}

	return ""
}

func (a args) GetValue(argNo int) string {
	source := os.Args[argNo]
	index := strings.Index(source, "=")

	if index == -1 {
		return ""
	}
	return source[index+1:]
}
