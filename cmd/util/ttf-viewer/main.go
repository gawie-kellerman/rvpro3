package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"rvpro3/radarvision.com/utils"
)

func main() {
	// 1. Load the TTF font
	width, height := 32, 32
	//fontBytes, err := os.ReadFile("/home/gkellerman/Workspace/RadarVision/Source/rvpro3/cmd/util/ttf-viewer/repet.ttf")
	//fontBytes, err := os.ReadFile("/home/gkellerman/Workspace/RadarVision/Source/rvpro3/cmd/util/ttf-viewer/lcd-7.ttf")
	fontBytes, err := os.ReadFile("/home/gkellerman/Downloads/ninepin.regular.ttf")
	utils.Debug.Panic(err)

	f, _ := truetype.Parse(fontBytes)

	// 2. Setup font face
	face := truetype.NewFace(f, &truetype.Options{
		Size:    8,
		DPI:     100,
		Hinting: font.HintingNone,
	})

	// Character to export
	char := 'W'

	// 3. Create a canvas for the bitmap
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// 4. Draw character
	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: face,
	}
	d.Dot = fixed.P(2, height-4) // Set baseline
	d.DrawString(string(char))

	for y := 0; y < height; y++ {
		fmt.Print(y, " ")
		for x := 0; x < width; x++ {
			pix := img.At(x, y)
			r, g, b, _ := pix.RGBA()

			if r == 0 && g == 0 && b == 0 {
				fmt.Print(" ")
			} else {
				fmt.Print("1")
			}
		}
		fmt.Println()
	}

	// 5. Save as PNG
	outFile, _ := os.Create("char_A.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
