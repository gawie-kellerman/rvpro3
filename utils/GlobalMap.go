package utils

type globalMap struct {
	data map[string]any
}

var GlobalMap globalMap

func init() {
	GlobalMap.data = make(map[string]any)
}

func (c *globalMap) Get(key string) any {
	return c.data[key]
}

func (c *globalMap) Set(key string, value any) {
	c.data[key] = value
}
