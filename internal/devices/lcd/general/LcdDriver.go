package general

import (
	"github.com/pkg/errors"
	"golang.org/x/exp/io/i2c"
	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
)

const callCommand = 0x00
const turnDisplayOff = 0xAE
const setStartLine = 0x40
const callColumnMapping = 0xA1
const comScanDec = 0xC8
const setComPins = 0xDA
const setContrast = 0x81
const setPreCharge = 0xD9
const setVComH = 0xDB
const setMemoryMode = 0x20
const displayAllOnResume = 0xA4
const normalDisplay = 0xA6
const turnDisplayOn = 0xAF
const setPageAddress = 0xB0
const setLowColumnAddr = 0x02
const setHighColumnAddr = 0x10
const callRam = 0x40

var ErrNotLCDDimensions = errors.New("not valid LCD dimensions")

type LcdDriver struct {
	handle *i2c.Device
}

func (l *LcdDriver) cmd(value byte) error {
	return l.handle.WriteReg(callCommand, []byte{value})
}

func (l *LcdDriver) cmdValue(command byte, value byte) error {
	_ = l.cmd(command)
	return l.cmd(value)
}

func (l *LcdDriver) Close() {
	if l.handle != nil {
		_ = l.handle.Close()
	}
}

func (l *LcdDriver) Open(device string, address int) (err error) {
	l.handle, err = i2c.Open(&i2c.Devfs{Dev: device}, address)
	if err != nil {
		return err
	}

	_ = l.cmd(turnDisplayOff)
	_ = l.cmd(setLowColumnAddr)
	_ = l.cmd(setHighColumnAddr)
	_ = l.cmd(setStartLine)
	_ = l.cmd(callColumnMapping)
	_ = l.cmd(comScanDec)

	_ = l.cmd(0xAD)
	_ = l.cmd(0x8B)
	_ = l.cmdValue(setComPins, 0x12)
	_ = l.cmdValue(setContrast, 0x80)
	_ = l.cmd(setPreCharge)
	_ = l.cmd(setVComH)
	_ = l.cmd(setStartLine)
	_ = l.cmdValue(setMemoryMode, 0x02)
	_ = l.cmd(displayAllOnResume)
	_ = l.cmd(normalDisplay)
	err = l.cmd(turnDisplayOn)

	return err
}

func (l *LcdDriver) DrawPage(content []byte, pageNo int) (err error) {
	err = l.cmd(setPageAddress + byte(pageNo))
	_ = l.cmd(setLowColumnAddr)
	_ = l.cmd(setHighColumnAddr)

	var buffer = [129]byte{}
	buffer[0] = callRam
	copy(buffer[1:], content)

	err = l.handle.Write(buffer[:])
	return err
}

func (l *LcdDriver) DrawMono(mono interfaces.IMonoBuffer) (err error) {
	if mono.Width() != 128 || mono.Height() != 64 {
		return ErrNotLCDDimensions
	}

	for page := 0; page < 8; page++ {
		if err = l.DrawPage(mono.GetPage(page), page); err != nil {
			return err
		}
	}
	return nil
}
