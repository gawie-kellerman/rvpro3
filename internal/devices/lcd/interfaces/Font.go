package interfaces

import (
	"encoding/binary"
	"fmt"

	"rvpro3/radarvision.com/utils/bit"
)

const firstCharU16Offset = 0
const lastCharU16Offset = 2
const charHeightU8Offset = 4
const fixedHeaderSize = charHeightU8Offset + 1

type Font struct {
	Definition []byte
	MinWidth   int
	MaxWidth   int
}

func (f *Font) GetFirstChar() uint16 {
	return binary.LittleEndian.Uint16(f.Definition[firstCharU16Offset : firstCharU16Offset+2])
}

func (f *Font) GetLastChar() uint16 {
	return binary.LittleEndian.Uint16(f.Definition[lastCharU16Offset : lastCharU16Offset+2])
}

func (f *Font) GetCharCount() uint16 {
	return f.GetLastChar() - f.GetFirstChar() + 1
}

func (f *Font) GetCharHeight() uint8 {
	return f.Definition[charHeightU8Offset]
}

func (f *Font) charOffsetLookup(absChar uint16) uint16 {
	relChar := absChar - f.GetFirstChar()
	u16Offset := fixedHeaderSize + relChar*2
	return binary.LittleEndian.Uint16(f.Definition[u16Offset : u16Offset+2])
}

func (f *Font) headerSize() uint16 {
	return fixedHeaderSize + uint16(f.GetCharCount())*2
}

func (f *Font) charOffset(absChar uint16) uint16 {
	off := f.charOffsetLookup(absChar)
	hs := f.headerSize()
	return off + hs
}

func (f *Font) GetCharWidth(ch uint16) uint8 {
	return f.Definition[f.charOffset(ch)]
}

func (f *Font) dataOffset(ch uint16, x, y int) uint16 {
	chOffset := f.charOffset(ch) + 1
	chBitWidth := f.GetCharWidth(ch)
	chByteWidth := bit.ByteWidth(chBitWidth)

	chOffset += uint16(y) * uint16(chByteWidth)
	chOffset += uint16(x / 8)
	return chOffset
}

func (f *Font) Is(ch uint16, x, y int) bool {
	chOffset := f.dataOffset(ch, x, y)
	chData := f.Definition[chOffset]

	return chData&(1<<(x%8)) != 0
}

func (f *Font) Get(ch uint16, x int, y int) bool {
	dataOffset := f.dataOffset(ch, x, y)
	data := f.Definition[dataOffset]
	return data&(1<<x%8) != 0
}

func (f *Font) DumpCh(ch uint16) {
	height := int(f.GetCharHeight())
	width := int(f.GetCharWidth(ch))

	fmt.Println(ch)
	fmt.Print("     ")
	for w := 0; w < width; w++ {
		fmt.Print(w % 10)
	}
	fmt.Println()

	for h := 0; h < height; h++ {
		fmt.Printf("%2d - ", h)
		for w := 0; w < width; w += 8 {
			byteValue := f.GetByte(ch, w, h)

			for iw := 0; iw < min(width-w, 8); iw++ {
				if byteValue&(1<<iw) == 0 {
					fmt.Print("_")
				} else {
					fmt.Print("0")
				}
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func (f *Font) DrawCh(buffer ILcdCanvas, ch uint16, x int, y int) int {
	height := int(f.GetCharHeight())
	width := int(f.GetCharWidth(ch))

	for h := 0; h < height; h++ {
		for w := 0; w < width; w += 8 {
			byteValue := f.GetByte(ch, w, h)

			for iw := 0; iw < min(width-w, 8); iw++ {
				plotX := x + iw + w
				plotY := h + y

				if byteValue&(1<<iw) == 0 {
					buffer.Clear(plotX, plotY)
				} else {
					buffer.Set(plotX, plotY)
				}
			}
		}
	}
	return width
}

func (f *Font) DrawStr(buffer ILcdCanvas, str string, x int, y int) int {
	runX := x

	for _, ch := range str {
		width := f.DrawCh(buffer, uint16(ch), runX, y)
		runX += width + 1

		if runX >= buffer.Width() {
			return runX
		}
	}
	return runX
}

func (f *Font) GetByte(ch uint16, w int, h int) uint8 {
	dataOffset := f.dataOffset(ch, w, h)
	return f.Definition[dataOffset]
}

func (f *Font) GetTextWidth(text string) int {
	runX := 0

	for _, ch := range text {
		runX += int(f.GetCharWidth(uint16(ch))) + 1
	}
	return runX
}
