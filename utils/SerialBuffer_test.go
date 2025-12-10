package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//goland:noinspection SpellCheckingInspection
const abc = "abcdefghij"
const startDelim = 'a'
const endDelim = 'j'

func TestDelimitBuffer_WithStrings(t *testing.T) {
	testString(t, "1234567890123", '1', '0', "1234567890", 10, 13)
	testString(t, "1234567890", '1', '0', "1234567890", 0, 0)
	testString(t, "1234567890", 'a', 'b', "", 9, 10)
}

func TestDelimitBuffer_Optimize(t *testing.T) {
	db := SerialBuffer{
		Buffer:     make([]byte, 100),
		StartDelim: startDelim,
		EndDelim:   endDelim,
	}

	for i := 0; i < 10; i++ {
		db.Push([]byte(abc))
	}

	for i := 0; i < 9; i++ {
		assert.Equal(t, abc, string(db.Pop()))
	}

	assert.Equal(t, 0, db.TailAvail())
	assert.Equal(t, 90, db.ReadPos)

	db.Optimize()

	assert.Equal(t, 90, db.TailAvail())
	assert.Equal(t, 0, db.ReadPos)
	assert.Equal(t, 10, db.Len())
}

func TestDelimitBuffer_PushOverflow(t *testing.T) {
	db := SerialBuffer{
		Buffer: make([]byte, 100),
	}

	for i := 0; i < 10; i++ {
		db.Push([]byte("abcdefghij"))
	}

	assert.Equal(t, ErrOverflow, db.Push([]byte("abcdefghij")))
}

func TestDelimitBuffer_Pop(t *testing.T) {
	db := SerialBuffer{
		Buffer:     make([]byte, 200),
		StartDelim: startDelim,
		EndDelim:   endDelim,
	}

	for i := 0; i < 10; i++ {
		db.Push([]byte(abc))
	}
	for i := 0; i < 10; i++ {
		assert.Equal(t, db.Pop(), abc)
	}

	assert.Equal(t, db.Pop(), "")
	assert.Equal(t, db.ReadPos, 0)
	assert.Equal(t, db.WritePos, 0)
}

func testString(t *testing.T, source string, startDelim byte, endDelim byte, expect string, readPos int, writePos int) {
	var buffer [256]byte

	db := SerialBuffer{
		Buffer: buffer[:],
	}

	Debug.Panic(db.Push([]byte(source)))
	actual := db.Pop()
	assert.Equal(t, expect, string(actual))
	assert.Equal(t, readPos, db.ReadPos, "readpos")
	assert.Equal(t, writePos, db.WritePos, "writepos")
}
