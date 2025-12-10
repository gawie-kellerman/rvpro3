package utils

import "github.com/pkg/errors"

var ErrOverflow = errors.New("SerialBuffer overflow")

// SerialBuffer is a highly optimized queue for writing and reading of delimited information
type SerialBuffer struct {
	Buffer     []byte
	ReadPos    int
	WritePos   int
	StartDelim byte
	EndDelim   byte
}

func (d *SerialBuffer) Len() int {
	return d.WritePos - d.ReadPos
}

func (d *SerialBuffer) Pop() []byte {
	rp := d.ReadPos
	wp := d.WritePos

	isFound := false

	for n := rp; n < wp; n++ {
		ch := d.Buffer[n]

		if ch == d.StartDelim {
			isFound = true
			d.ReadPos = n
		} else if ch == d.EndDelim {
			if isFound {
				res := d.Buffer[d.ReadPos : n+1]
				d.ReadPos = n + 1

				if d.ReadPos+1 >= d.WritePos {
					d.Reset()
				}
				return res
			}
		} else {
			if !isFound {
				d.ReadPos = n
			}
		}

	}

	return nil
}

func (d *SerialBuffer) Push(data []byte) error {
	dataLen := len(data)
	if d.ShouldOptimize(dataLen) {
		d.Optimize()
	}

	if !d.WillFit(dataLen) {
		return ErrOverflow
	}

	copy(d.Buffer[d.WritePos:], data)
	d.WritePos += dataLen
	return nil
}

func (d *SerialBuffer) ShouldOptimize(size int) bool {
	if d.TailAvail() >= size {
		return false
	}

	if d.TotalAvail() >= size {
		return true
	}

	return false
}

func (d *SerialBuffer) Reset() {
	d.WritePos = 0
	d.ReadPos = 0
}

func (d *SerialBuffer) TailAvail() int {
	return cap(d.Buffer) - d.WritePos
}

func (d *SerialBuffer) HeadAvail() int {
	return d.ReadPos
}

func (d *SerialBuffer) TotalAvail() int {
	return d.TailAvail() + d.ReadPos
}

func (d *SerialBuffer) WillFit(size int) bool {
	return d.TailAvail() >= size
}

func (d *SerialBuffer) Optimize() {
	copy(d.Buffer[0:], d.Buffer[d.ReadPos:d.WritePos])
	d.WritePos = d.WritePos - d.ReadPos
	d.ReadPos = 0
}
