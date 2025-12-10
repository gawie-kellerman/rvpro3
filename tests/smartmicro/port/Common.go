package port

import "path"

func getT45Path(filename string) string {
	return path.Join("t45/20251128/", filename)
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}
