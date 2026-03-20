package general

import (
	"fmt"
	"image"
)

type LcdCanvas struct {
	width  int
	height int
	Buffer []byte
}

func (l *LcdCanvas) Init(width int, height int) {
	l.width = width
	l.height = height
	l.Buffer = make([]byte, l.ByteTotal())
}

func (l *LcdCanvas) InitSlice(width int, height int, slice []byte) {
	l.width = width
	l.height = height
	l.Buffer = slice
}

func (l *LcdCanvas) ByteWidth() int {
	return (l.width + (8 - 1)) / 8
}

func (l *LcdCanvas) ByteTotal() int {
	return l.height * l.ByteWidth()
}

func (l *LcdCanvas) ByteOffset(x int, y int) int {
	return y*l.ByteWidth() + (x / 8)
}

func (*LcdCanvas) BitOffset(x int) int {
	return x % 8
}

func (l *LcdCanvas) Offsets(x, y int) (int, byte) {
	return x + (y/8)*l.width, 1 << (y % 8)
}

func (l *LcdCanvas) Set(x int, y int) {
	offset, pixel := l.Offsets(x, y)

	l.Buffer[offset] |= pixel
}

func (l *LcdCanvas) Clear(x int, y int) {
	if x < 0 || x >= l.width {
		return
	}
	if y < 0 || y >= l.height {
		return
	}

	offset, pixel := l.Offsets(x, y)

	l.Buffer[offset] &= ^pixel
}

func (l *LcdCanvas) Get(x int, y int) bool {
	byteOffset := l.ByteOffset(x, y)
	bitOffset := l.BitOffset(x)
	mask := uint8(1 << bitOffset)

	return (l.Buffer[byteOffset] & mask) == mask
}

func (l *LcdCanvas) Is(x, y int) bool {
	byteOffset := l.ByteOffset(x, y)
	bitOffset := l.BitOffset(x)
	return (l.Buffer[byteOffset] & (1 << bitOffset)) == (1 << bitOffset)
}

func (l *LcdCanvas) GetBuffer(fromLine int) []byte {
	offset := l.ByteOffset(0, fromLine)
	return l.Buffer[offset:]
}

func (l *LcdCanvas) DrawYOffset(yOffset int, mask image.Image) {
	for y := 0; y < mask.Bounds().Dy(); y++ {
		for x := 0; x < mask.Bounds().Dx(); x++ {
			col := mask.At(x, y)
			r, g, b, a := col.RGBA()
			if r == 0 || g == 0 || b == 0 || a == 0 {
			} else {
				l.Set(x, y+yOffset)
			}
		}
	}
}

func (l *LcdCanvas) HorizontalLine(y int) {
	for x := 0; x < l.width; x++ {
		l.Set(x, y)
	}
}

func (l *LcdCanvas) DumpToConsole() {
	fmt.Print("   - ")
	for x := 0; x < l.width; x++ {
		fmt.Print(x % 10)
	}
	fmt.Println()

	for y := 0; y < l.height; y++ {
		fmt.Printf("%02d - ", y)
		for x := 0; x < l.width; x++ {
			if l.Get(x, y) {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

}

func (l *LcdCanvas) Copy(cur []byte, start int) {
	copy(l.Buffer[start:len(cur)], cur)
}

func (l *LcdCanvas) GetPage(pageNo int) []byte {
	return l.Buffer[pageNo*l.width : pageNo*l.width+l.width]
}

func (l *LcdCanvas) GetPageStart(no int) int {
	return l.width * no
}

func (l *LcdCanvas) Width() int {
	return l.width
}

func (l *LcdCanvas) Height() int {
	return l.height
}

func (l *LcdCanvas) Fill(pattern byte) {
	for n := 0; n < len(l.Buffer); n++ {
		l.Buffer[n] = pattern
	}
}

func (l *LcdCanvas) DrawText(font *Font, text string, x int, y int) {
	font.DrawStr(l, text, x, y)
}
