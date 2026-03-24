package interfaces

type ILcdCanvas interface {
	Set(x, y int)
	Clear(x, y int)
	Get(x, y int) bool
	Is(x, y int) bool

	SetXY(x, y int) ILcdCanvas

	DrawStrLn(text string) ILcdCanvas
	DrawStr(text string) ILcdCanvas
	DrawSymbol(font *Font, symbol uint16) ILcdCanvas
	DrawRight(right int, format string)
	DrawLn()

	MoveCursorBy(x, y int) ILcdCanvas
	MoveLn()

	ClearScreen()
	HorzLine(y int) ILcdCanvas
	SetX(i int) ILcdCanvas

	GetY() int

	Width() int
	Height() int

	ByteWidth() int
	ByteTotal() int
	ByteOffset(x int, y int) int
	BitOffset(x int) int

	GetPage(pageNo int) []byte

	Fill(pattern byte)

	SetFont(font *Font)
	GetFont() *Font
}
