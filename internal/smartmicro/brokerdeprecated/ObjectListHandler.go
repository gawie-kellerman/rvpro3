package brokerdeprecated

import (
	"fmt"

	"rvpro3/radarvision.com/internal/smartmicro/port"
)

type ObjectListHandler struct{}

func (s *ObjectListHandler) Handle(msg *Message) {
	var objList port.ObjectList
	var err error

	if err = objList.ReadBytes(msg.Data[:msg.DataLen]); err != nil {
		s.logMappingErr()
	}

	fmt.Println("Object List with :", objList.Header.NofObjects, " objects")
}

func (s *ObjectListHandler) logMappingErr() {
	fmt.Println("Error reading object mapping")
}
