package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type mapSerializer struct{}

func (mapSerializer) Load(filename string) (map[string]string, error) {
	res := make(map[string]string)

	parseLine := func(line string) {
		index := strings.Index(line, "=")

		if index == -1 {
			return
		}

		key := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])

		res[key] = value
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parseLine(line)
	}
	return res, scanner.Err()
}

func (mapSerializer) Save(data map[string]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()

	for key, value := range data {
		_, err = fmt.Fprintln(w, key, "=", value)
		if err != nil {
			return err
		}
	}
	return nil
}

var MapSerializer = mapSerializer{}
