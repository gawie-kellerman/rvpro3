package interfaces

type IMonoBuffer interface {
	Set(x, y int)
	Clear(x, y int)
	Get(x, y int) bool
	Is(x, y int) bool

	Width() int
	Height() int

	ByteWidth() int
	ByteTotal() int
	ByteOffset(x int, y int) int
	BitOffset(x int) int

	GetPage(pageNo int) []byte

	Fill(pattern byte)

	HorizontalLine(y int)
}
