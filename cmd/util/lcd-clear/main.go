package main

import (
	"fmt"
	"strings"

	"rvpro3/radarvision.com/internal/devices/lcd/fonts"
	"rvpro3/radarvision.com/internal/devices/lcd/general"
	"rvpro3/radarvision.com/utils"
)

func main() {

	const device = "/dev/i2c-1"
	const address = 0x3c

	utils.Print.Ln("Device:", device, "Address:", fmt.Sprintf("0x%x", address))
	utils.Print.Ln("Opening I2C/LCD Driver")

	driver := general.LcdDriver{}
	utils.Debug.Panic(driver.Open(device, address))
	defer driver.Close()

	mono := general.LcdCanvas{}
	mono.Init(128, 64)

	cmd := strings.ToLower(utils.Args.Command(1, "clear"))
	switch cmd {
	case "fonts":
		y := 0
		//fonts.TextA8D2.DrawStr(&mono, "Text A8D2 1234567890:[*]", 0, y)
		//y += int(fonts.TextA8D2.GetCharHeight()) + 2

		//fonts.TextA5D1.DrawStr(&mono, "Text A5D1 1234567890:[*]", 0, y)
		//y += int(fonts.TextA5D1.GetCharHeight()) + 2
		//
		//fonts.RobotoA10D2.DrawStr(&mono, "Text Roboto 1234567890:[*]", 0, y)
		//y += int(fonts.RobotoA10D2.GetCharHeight()) + 2

		fonts.WingdingsA5D1.DrawStr(&mono, string([]byte{32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53}), 0, y)
		y += int(fonts.WingdingsA5D1.GetCharHeight()) + 2

		fonts.NotoSansA6D2.DrawStr(&mono, "Text NotoSans 1234567890:[*]", 0, y)
		y += int(fonts.NotoSansA6D2.GetCharHeight()) + 2

		fonts.IBMPlexA8D3.DrawStr(&mono, "Text IBMPlex 1234567890:[*]", 0, y)
		y += int(fonts.NotoSansA6D2.GetCharHeight()) + 2

		fonts.VerdanaA8D1.DrawStr(&mono, "Text Verdana 1234567890:[*]", 0, y)
		y += int(fonts.VerdanaA8D1.GetCharHeight()) + 2

		fonts.CozetteA10D3.DrawStr(&mono, "Text Cozette A10 1234567890:[*]", 0, y)
		y += int(fonts.CozetteA10D3.GetCharHeight()) + 2

		//fonts.TextA5D1.DrawStr(&mono, "ABCDEF", 0, 0)
		//fonts.RobotoMonoA7D1.DrawStr(&mono, "ROBOTO MONO!", 0, 0)
		//mono.DrawText(fonts.TextA8D2, "Hello world!", 0, 0)
		mono.HorizontalLine(32)
	case "fill":
		utils.Print.Ln("Filling the LCD")
		mono.Fill(0xff)

	default:
		utils.Print.Ln("Clearing the LCD")
		mono.Fill(0x00)
	}

	utils.Debug.Panic(driver.DrawMono(&mono))
	utils.Print.Ln("Done")
}
