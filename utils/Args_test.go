package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgs_IndexOf(t *testing.T) {
	os.Args = append(os.Args, "--filename=Hungry Hippo.txt")

	fmt.Println(os.Args)

	index, argName := Args.IndexOf("--filename|-f")
	fmt.Println(index, argName)
	assert.NotEqual(t, -1, index)

	assert.Equal(t, "Hungry Hippo.txt", Args.GetValue(index))
}

func TestArgs_KeyValue(t *testing.T) {
	os.Args = append(os.Args, "-oShould.Print=true")
	os.Args = append(os.Args, "-oShould.Print=false")
	os.Args = append(os.Args, "-oShould.Flag")

	const pn = "-o|--override"
	indexes := Args.GetKVPairIndexes(pn)
	for _, index := range indexes {
		fmt.Println("Name", Args.GetKeyName(index, pn), "Value", Args.GetValue(index))
	}
}

func TestHas_Flag(t *testing.T) {
	os.Args = append(os.Args, "--Should.Print")

	assert.Equal(t, true, Args.Has("--Should.Print"))
}

func TestArgs_Default(t *testing.T) {
	os.Args = append(os.Args, "--Should.Print")

	res := Args.GetString("--Should.Print=", "Default Value")
	assert.Equal(t, "Default Value", res)
}
