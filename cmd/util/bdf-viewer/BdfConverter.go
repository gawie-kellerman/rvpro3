package main

import (
	"fmt"
	"os"

	"github.com/tidbyt/go-bdf"
	"golang.org/x/image/math/fixed"
)

type bdfConverter struct{}

var BdfConverter = bdfConverter{}

func (c *bdfConverter) Convert(
	baseName string,
	filePath string,
	startChar uint16,
	endChar uint16,
) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	fnt, err := bdf.Parse(bytes)
	if err != nil {
		return "", err
	}

	height := fnt.Ascent + fnt.Descent
	//width := fnt.PixelSize

	face := fnt.NewFace()

	defBuilder := FontDefinitionBuilder{}
	defBuilder.MinWidth = 100

	fontName := fmt.Sprintf("%sA%dD%d", baseName, fnt.Ascent, fnt.Descent)

	defBuilder.Init(fontName, startChar, endChar, uint8(height))

	dot := fixed.Point26_6{X: 0, Y: face.Metrics().Ascent}

	for n := defBuilder.startChar; n <= defBuilder.endChar; n++ {
		fmt.Println(n)
		rect, mask, _, _, _ := face.Glyph(dot, rune(n))
		yOffset := 0
		if mask.Bounds().Dy() <= height {
			yOffset = rect.Min.Y
			if yOffset+mask.Bounds().Dy() > height {
				yOffset = height - mask.Bounds().Dy()
			}
			yOffset = max(0, yOffset)
		}

		charWidth := mask.Bounds().Dx()
		defBuilder.MinWidth = min(defBuilder.MinWidth, charWidth)
		defBuilder.MaxWidth = max(defBuilder.MaxWidth, charWidth)
		if defBuilder.startChar+1 >= n {
			//defBuilder.AddFullTest(0, 8, height)
			//defBuilder.AddBorderTest(0, 8, height)
			defBuilder.Add(yOffset, height, mask)
		} else {
			defBuilder.Add(yOffset, height, mask)
		}
	}
	return defBuilder.String(), nil
}
