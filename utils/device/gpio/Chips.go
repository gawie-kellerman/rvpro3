package gpio

import (
	"github.com/warthog618/go-gpiocdev"
)

type Chips struct {
	Chips map[string]*Chip
}

func (c *Chips) Init() {
	c.Chips = make(map[string]*Chip)
}

func (c *Chips) GetChip(name string) *Chip {
	return c.Chips[name]
}

func (c *Chips) CloseChip(name string) {
	chip, ok := c.Chips[name]
	if ok {
		chip.Close()
		delete(c.Chips, name)
	}
}

// OpenByChip creates a new gpiocdev.Chip object.  Close will close all the Chip
func (c *Chips) OpenByChip(chipNo int, options ...gpiocdev.ChipOption) (*Chip, error) {
	chipName := Util.GetChipName(chipNo)
	chip, ok := c.Chips[chipName]
	if ok {
		return chip, nil
	}
	chipObj, err := gpiocdev.NewChip(chipName, options...)

	if err != nil {
		return nil, err
	}

	chip = new(Chip)
	chip.Init(chipNo, chipObj)

	c.Chips[chipName] = chip

	return chip, nil
}

func (c *Chips) OpenByPort(port int, options ...gpiocdev.ChipOption) (*Chip, error) {
	return c.OpenByChip(Util.GetChipNo(port), options...)
}

func (c *Chips) Close() {
	for _, chip := range c.Chips {
		chip.Close()
	}
	clear(c.Chips)
}
