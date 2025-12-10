package utils

type args struct {
	data map[string]any
}

var Args args

func init() {
	Args.data = make(map[string]any)
}

func (c *args) Get(key string) any {
	return c.data[key]
}

func (c *args) Set(key string, value any) {
	c.data[key] = value
}
