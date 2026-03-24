package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type basic struct {
	data map[string]string
}

func (b *basic) getKey(key string) string {
	return key
}

func (b *basic) Set(key string, value string) {
	fullKey := b.getKey(key)
	b.data[fullKey] = value
}

func (b *basic) Get(key string, def string) string {
	globalKey := b.getKey(key)
	if res, ok := b.data[globalKey]; ok {
		return res
	}
	b.data[globalKey] = def
	return def
}

func (b *basic) GetArray(key string, def string) []string {
	strArr := b.Get(key, def)
	return strings.Split(strArr, ";")
}

func (b *basic) SetBool(key string, value bool) {
	strValue := strconv.FormatBool(value)
	b.Set(key, strValue)
}

func (b *basic) logErr(key string, err error, msg string) {
	log.Err(err).Str("key", key).Msg(msg)
}

func (b *basic) GetBool(key string, defValue bool) bool {
	value := b.Get(key, strconv.FormatBool(defValue))

	res, err := strconv.ParseBool(value)
	if err != nil {
		b.logErr(b.getKey(key), err, "Error parsing bool value")
	}
	return res
}

func (b *basic) SetInt(key string, value int) {
	strValue := strconv.Itoa(value)
	b.Set(key, strValue)
}

func (b *basic) GetInt(key string, defValue int) int {
	value := b.Get(key, strconv.Itoa(defValue))

	res, err := ParseInt(value, defValue)
	if err != nil {
		b.logErr(b.getKey(key), err, "Error parsing int value")
	}
	return res
}

func (b *basic) GetMilliseconds(key string, defValue int) Milliseconds {
	value := b.GetInt(key, defValue)
	return Milliseconds(time.Duration(value) * time.Millisecond)
}

func (b *basic) Has(key string) bool {
	globalKey := b.getKey(key)
	_, ok := b.data[globalKey]

	return ok
}

func (b *basic) GetIP4(key string, ip4 IP4) IP4 {
	value := b.Get(key, ip4.String())

	return IP4Builder.FromString(value)
}
