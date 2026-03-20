package main

import (
	"fmt"
	"image"
	"strings"

	"rvpro3/radarvision.com/utils/bit"
)

type FontDefinitionBuilder struct {
	header            strings.Builder
	data              strings.Builder
	headerCount       int
	dataCount         int
	startChar         int
	endChar           int
	currentChar       int
	currentCharOffset uint16
	MinWidth          int
	MaxWidth          int
	offset            int
}

func (fb *FontDefinitionBuilder) Init(name string, startChar uint16, endChar uint16, charHeight uint8) {
	fb.header.WriteString("var ")
	fb.header.WriteString(name)
	fb.header.WriteString(" = ")
	fb.header.WriteString("&general.Font{\n")
	fb.header.WriteString("    Definition: []byte{\n")

	fb.header.WriteString("        ")
	fb.U16(&fb.header, startChar)
	fb.U16(&fb.header, endChar)
	fb.header.WriteString(fmt.Sprintf("0x%02X, \n        ", charHeight))

	fb.startChar = int(startChar)
	fb.endChar = int(endChar)
	fb.currentChar = fb.startChar
}

func (fb *FontDefinitionBuilder) String() string {
	fb.done()
	fb.header.WriteString("// End Of Header\n        ")
	fb.header.WriteString(fb.data.String())
	fb.header.WriteString("\n")
	return fb.header.String()
}

//func (fb *FontDefinitionBuilder) Add(width uint8, charBuffer []byte) {
//	utils.Debug.PanicIf(fb.currentChar > fb.endChar, "Too many characters")
//
//	fb.HeaderU8(uint8(fb.currentCharOffset & 0xFF))
//	fb.HeaderU8(uint8(fb.currentCharOffset >> 8))
//
//	fb.DataU8(width)
//	for _, ch := range charBuffer {
//		fb.DataU8(ch)
//	}
//
//	fb.currentCharOffset += uint16(len(charBuffer) + 1)
//	fb.currentChar += 1
//}

func (fb *FontDefinitionBuilder) done() {
	fb.data.WriteString("\n    },\n")
	fb.data.WriteString(fmt.Sprintf("    MinWidth: %d, \n", fb.MinWidth))
	fb.data.WriteString(fmt.Sprintf("    MaxWidth: %d, \n", fb.MaxWidth))
	fb.data.WriteString("}\n")
}

func (fb *FontDefinitionBuilder) HeaderU8(value uint8) {
	fb.headerCount++

	if fb.headerCount%16 == 0 {
		//fb.headerCount = 0
		fb.header.WriteString("\n        ")
	}

	fb.header.WriteString(fmt.Sprintf("0x%02X, ", value))
}

func (fb *FontDefinitionBuilder) DataU8(value uint8) {
	fb.dataCount++

	if fb.dataCount%16 == 0 {
		//fb.dataCount = 0
		fb.data.WriteString("\n        ")
	}

	fb.data.WriteString(fmt.Sprintf("0x%02X, ", value))
}

func (fb *FontDefinitionBuilder) U16(builder *strings.Builder, value uint16) {
	builder.WriteString(fmt.Sprintf("0x%02X, ", value&0xFF))
	builder.WriteString(fmt.Sprintf("0x%02X, ", value>>8))
}

func (fb *FontDefinitionBuilder) AddFullTest(yOffset int, dx int, dy int) {
	width := dx
	byteWidth := bit.ByteWidth(width)

	fb.HeaderU16(fb.offset)

	offset := 1
	fb.DataU8(uint8(dx))

	// Leading empty data
	for _ = range yOffset * byteWidth {
		fb.DataU8(0)
		offset += 1
	}

	var value uint8
	var ch bool

	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			ch = true
			r, g, b, a := 1, 1, 1, 1
			if r == 0 && g == 0 && b == 0 && a == 0 {
				// Do nothing
			} else {
				value |= 1 << (x % 8)
			}

			if x > 0 && x%8 == 0 {
				fb.DataU8(value)
				offset += 1
				value = 0
				ch = false
			}
		}
		if ch {
			fb.DataU8(value)
			offset += 1
			value = 0
		}
	}
	fb.offset += offset

}

func (fb *FontDefinitionBuilder) AddBorderTest(yOffset int, dx int, dy int) {
	width := dx
	byteWidth := bit.ByteWidth(width)

	fb.HeaderU16(fb.offset)

	offset := 1
	fb.DataU8(uint8(dx))

	// Leading empty data
	for _ = range yOffset * byteWidth {
		fb.DataU8(0)
		offset += 1
	}

	var value uint8
	var ch bool

	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			ch = true
			r, g, b, a := 0, 0, 0, 0
			if y == 0 || x == 0 || y == dy-1 || x == dx-1 {
				r = 1
			}

			if r == 0 && g == 0 && b == 0 && a == 0 {
				// Do nothing
			} else {
				value |= 1 << (x % 8)
			}

			if x > 0 && x%8 == 0 {
				fb.DataU8(value)
				offset += 1
				value = 0
				ch = false
			}
		}
		if ch {
			fb.DataU8(value)
			offset += 1
			value = 0
		}
	}
	fb.offset += offset

}

func (fb *FontDefinitionBuilder) Add(yOffset int, yHeight int, mask image.Image) {
	width := mask.Bounds().Dx()
	height := mask.Bounds().Dy()
	byteWidth := bit.ByteWidth(width)

	fb.HeaderU16(fb.offset)

	offset := 1
	fb.DataU8(uint8(mask.Bounds().Dx()))

	// Leading empty data
	for _ = range yOffset * byteWidth {
		fb.DataU8(0)
		offset += 1
	}

	yRunner := yOffset

	var value uint8
	var ch bool

	for y := 0; y < height; y++ {
		yRunner += 1

		for x := 0; x < width; x++ {
			ch = true

			r, g, b, a := mask.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				// Do nothing
			} else {
				value |= 1 << (x % 8)
			}

			if x > 0 && x%8 == 0 {
				fb.DataU8(value)
				offset += 1
				value = 0
				ch = false
			}
		}

		if ch {
			fb.DataU8(value)
			offset += 1
			value = 0
		}
	}

	for ; yRunner < yHeight; yRunner++ {
		for _ = range byteWidth {
			fb.DataU8(0)
			offset += 1
		}
	}

	fb.offset += offset
}

func (fb *FontDefinitionBuilder) HeaderU16(value int) {
	fb.HeaderU8(uint8(value & 0xFF))
	fb.HeaderU8(uint8(value >> 8))
}
