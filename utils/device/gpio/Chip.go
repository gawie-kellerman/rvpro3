package gpio

import "github.com/warthog618/go-gpiocdev"

type Chip struct {
	No           int
	Chip         *gpiocdev.Chip
	Lines        map[int]*gpiocdev.Line
	OnHandleGPIO func(gpio *Chip, event gpiocdev.LineEvent)
}

func (c *Chip) EventHandler(event gpiocdev.LineEvent) {
	if c.OnHandleGPIO != nil {
		c.OnHandleGPIO(c, event)
	}
}

func (c *Chip) Init(no int, chip *gpiocdev.Chip) {
	c.No = no
	c.Chip = chip
	c.Lines = make(map[int]*gpiocdev.Line)
}

func (c *Chip) Close() {
	for _, line := range c.Lines {
		_ = line.Close()
	}
	clear(c.Lines)

	_ = c.Chip.Close()
}

func (c *Chip) ListenToLine(offset int) (line *gpiocdev.Line, err error) {
	var ok bool

	line, ok = c.Lines[offset]
	if ok {
		return line, nil
	}

	line, err = c.Chip.RequestLine(
		offset,
		gpiocdev.WithEventHandler(c.EventHandler),
		gpiocdev.LineEdgeBoth,
	)
	if err != nil {
		return nil, err
	}
	c.Lines[offset] = line
	return line, nil
}

func (c *Chip) WriteToLine(offset int, hi int, lo int) (line *gpiocdev.Line, err error) {
	var ok bool

	line, ok = c.Lines[offset]
	if ok {
		return line, nil
	}

	line, err = c.Chip.RequestLine(offset, gpiocdev.AsOutput(hi, lo))
	if err != nil {
		return nil, err
	}
	c.Lines[offset] = line
	return line, nil
}
