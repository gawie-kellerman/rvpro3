package utils

import (
	"encoding/hex"
	"os"
)

var File file

type file struct {
}

func (file) LoadFromHex(filename string) (res []byte, err error) {
	if res, err = os.ReadFile(filename); err != nil {
		return nil, err
	}

	if res, err = hex.DecodeString(string(res)); err != nil {
		return nil, err
	}

	return res, nil
}

func (f file) SaveAsHex(filename string, slice []byte) error {
	encoded := hex.EncodeToString(slice)
	return os.WriteFile(filename, []byte(encoded), 0644)
}

func (f file) Exists(filename string) (bool, error) {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
