package main

import (
	"fmt"
	"os"

	"github.com/tidbyt/go-bdf"
	"golang.org/x/image/math/fixed"
	"rvpro3/radarvision.com/utils"
)

//https://www.nerdfonts.com/cheat-sheet

const IconEmptyCircle = 33
const IconFullCircle = 34
const IconStruckCircle = 35
const IconRight = 36
const IconLeft = 37
const IconUp = 38
const IconDown = 39
const IconEmptyBox = 40
const IconOBox = 41
const IconStruckBox = 42
const IconCheckedBox = 43
const IconArrowRight = 44
const IconArrowLeft = 45
const IconArrowUp = 46
const IconArrowDown = 47
const IconProgress1 = 48
const IconProgress2 = 49
const IconProgress3 = 50
const IconProgress4 = 51
const IconProgressDot = 52
const IconConnected = 52
const IconDisconnected = 53

func main() {
	createOfficialFonts()
}

func createOfficialFonts() {
	createOfficialFont("Wingdings", "/home/gkellerman/Workspace/RadarVision/Source/rvpro3/cmd/util/bdf-viewer/fonts/Wingdings_5pt-7.bdf", 32, 54)
	createOfficialFont("Roboto", "/home/gkellerman/Downloads/bdf/fonts/RobotoMono-Regular-7pt.bdf", 32, 126)
	createOfficialFont("NotoSans", "/home/gkellerman/Downloads/bdf/fonts/NotoSans-Regular-8px.bdf", 32, 126)
	createOfficialFont("Verdana", "/home/gkellerman/Downloads/bdf/fonts/verdana-6pt.bdf", 32, 126)
	//createOfficialFont("NotoSans", "/home/gkellerman/Downloads/bdf/fonts/04B_03__7pt.bdf", 32, 126)
}

func createOfficialFont(baseName string, filePath string, startChar uint16, endChar uint16) {
	fontSrc, err := BdfConverter.Convert(baseName, filePath, startChar, endChar)
	utils.Debug.Panic(err)
	utils.Print.RawLn(fontSrc)
}

func doPicker() {
	picker := BdfPicker{}
	utils.Debug.Panic(picker.Init("/home/gkellerman/Downloads/out.bdf", 32, 41))

	picker.Add(62099)
	picker.Add(61796)
	picker.Add(60082)
	picker.Add(0xf1eb)
	picker.Add(0xee06)
	picker.Add(0xee07)
	picker.Add(0xee08)
	picker.Add(0xee09)
	picker.Add(0xee0a)
	picker.Add(0xee0b)
	//picker.Add(0xf16c6)
	fmt.Println(picker.String())
}

func oldMain() {
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/Roboto-Regular-7pt.bdf") GOOD
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/Roboto-Regular-7pt.bdf") GOOD
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/hyundai-led-8pt.bdf") BAD
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/l_10646-6pt.bdf") ALRIGHT
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/NotoSans-Regular-8px.bdf") GOOD, BROKEN [
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/MartianMono_Condensed-Regular-6px.bdf") ONLY UPPERS
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/IBMPlexMono-Regular-6px.bdf")  TOO SMALL
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/04B_03__7pt.bdf") BEST
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/bdf/fonts/04B_03__6pt.bdf") BEST SMALL
	//https://github.com/IT-Studio-Rech/bdf-fonts
	bytes, err := os.ReadFile("/home/gkellerman/Downloads/6x10.bdf")
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/6x9.bdf")
	//bytes, err := os.ReadFile("/home/gkellerman/Downloads/HaxorMedium-10.bdf") NICE

	utils.Debug.Panic(err)

	fnt, err := bdf.Parse(bytes)
	utils.Debug.Panic(err)

	face := fnt.NewFace()

	var matrix [][]bool
	height := fnt.Ascent + fnt.Descent
	width := fnt.PixelSize

	matrix = make([][]bool, height)
	for i := range height {
		matrix[i] = make([]bool, width)
	}

	//dot := fixed.Point26_6{X: 0, Y: 0}
	dot := fixed.Point26_6{X: 0, Y: face.Metrics().Ascent}

	rect, mask, _, _, ok := face.Glyph(dot, 'A')
	utils.Debug.PanicIf(!ok, "Glyph")

	dx := mask.Bounds().Dx()
	dy := mask.Bounds().Dy()

	utils.Debug.PanicIf(width < dx, "x")
	utils.Debug.PanicIf(height < dy, "y")

	for y := 0; y < mask.Bounds().Dy(); y++ {
		for x := 0; x < mask.Bounds().Dx(); x++ {
			col := mask.At(x, y)
			r, g, b, a := col.RGBA()
			if r == 0 || g == 0 || b == 0 || a == 0 {
			} else {
				yOff := y + (rect.Min.Y)
				matrix[yOff][x] = true
			}
		}
	}

	for y := range matrix {
		fmt.Printf("%2d - ", y)
		for x := range matrix[y] {
			if matrix[y][x] {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	//fmt.Println(rect, mask, pt, advance, ok)
	//
	//fmt.Print("   - ")
	//for x := 0; x < mask.Bounds().Dx(); x++ {
	//	fmt.Print(strconv.Itoa(x % 10))
	//}
	//fmt.Println()
	//for y := 0; y < mask.Bounds().Dy(); y++ {
	//	fmt.Printf("%2d - ", y)
	//	for x := 0; x < mask.Bounds().Dx(); x++ {
	//		col := mask.At(x, y)
	//		r, g, b, a := col.RGBA()
	//		if r == 0 || g == 0 || b == 0 || a == 0 {
	//			fmt.Print(" ")
	//		} else {
	//			fmt.Print("1")
	//		}
	//	}
	//	fmt.Println()
	//}
}
