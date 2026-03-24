package general

import (
	"fmt"
	"image"

	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
)

type LcdCanvas struct {
	width  int
	height int
	Buffer []byte
	Font   *interfaces.Font
	X      int
	Y      int
}

func (l *LcdCanvas) Init(width int, height int, font *interfaces.Font) {
	l.width = width
	l.height = height
	l.Buffer = make([]byte, l.ByteTotal())
	l.Font = font
}

func (l *LcdCanvas) SetFont(font *interfaces.Font) {
	l.Font = font
}

func (l *LcdCanvas) GetFont() *interfaces.Font {
	return l.Font
}

func (l *LcdCanvas) SetXY(x, y int) interfaces.ILcdCanvas {
	l.X = x
	l.Y = y
	return l
}

func (l *LcdCanvas) SetX(x int) interfaces.ILcdCanvas {
	l.X = x
	return l
}

func (l *LcdCanvas) DrawStrLn(text string) interfaces.ILcdCanvas {
	_ = l.Font.DrawStr(l, text, l.X, l.Y)
	return l.MoveCursorBy(0, int(l.Font.GetCharHeight()))
}

func (l *LcdCanvas) DrawLn() {
	l.MoveCursorBy(0, int(l.Font.GetCharHeight()))
}

func (l *LcdCanvas) MoveLn() {
	l.Y += int(l.Font.GetCharHeight())
	l.X = 0
}

func (l *LcdCanvas) DrawRight(right int, text string) {
	x := l.Font.GetTextWidth(text)
	l.Font.DrawStr(l, text, right-x, l.Y)
	l.X = right
}

func (l *LcdCanvas) DrawStr(text string) interfaces.ILcdCanvas {
	width := l.Font.DrawStr(l, text, l.X, l.Y)
	return l.MoveCursorBy(width, 0)
}

func (l *LcdCanvas) DrawSymbol(font *interfaces.Font, symbol uint16) interfaces.ILcdCanvas {
	width := font.DrawCh(
		l,
		symbol,
		l.X,
		l.Y,
	)

	return l.MoveCursorBy(width, 0)
}

func (l *LcdCanvas) GetY() int {
	return l.Y
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
	if x < 0 || x >= l.width {
		return
	}
	if y < 0 || y >= l.height {
		return
	}
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

func (l *LcdCanvas) HorzLine(y int) interfaces.ILcdCanvas {
	for x := 0; x < l.width; x++ {
		l.Set(x, y)
	}
	return l
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
	copy(l.Buffer[start:start+len(cur)], cur)
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

func (l *LcdCanvas) DrawText(font *interfaces.Font, text string, x int, y int) {
	font.DrawStr(l, text, x, y)
}

func (l *LcdCanvas) MoveCursorBy(x int, y int) interfaces.ILcdCanvas {
	l.X += x
	l.Y += y
	return l
}

func (l *LcdCanvas) ClearScreen() {
	clear(l.Buffer)
	l.X, l.Y = 0, 0
}
