package main

import (
	"os"

	"github.com/tidbyt/go-bdf"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type BdfPicker struct {
	buffer   FontDefinitionBuilder
	font     *bdf.Font
	Filename string
	face     font.Face
	dot      fixed.Point26_6
	height   uint8
}

func (b *BdfPicker) String() string {
	return b.buffer.String()
}

func (b *BdfPicker) Init(filename string, startChar uint16, endChar uint16) (err error) {
	var bytes []byte
	b.Filename = filename

	if bytes, err = os.ReadFile(filename); err != nil {
		return err
	}

	if b.font, err = bdf.Parse(bytes); err != nil {
		return err
	}

	b.height = uint8(b.font.Ascent + b.font.Descent)

	b.face = b.font.NewFace()
	b.buffer.Init("name", startChar, endChar, b.height)
	b.dot = fixed.Point26_6{X: 0, Y: b.face.Metrics().Ascent}
	return nil
}

//func (b *BdfPicker) add(char rune) {
//	rect, mask, _, _, _ := b.face.Glyph(b.dot, rune(char))
//
//	mono := general.LcdCanvas{}
//	mono.Init(b.font.PixelSize, int(b.height), b.font)
//	mono.DrawYOffset(rect.Min.Y, mask)
//	mono.DumpToConsole()
//
//	b.buffer.Add(mono.Width(), 8, mask)
//}
//
