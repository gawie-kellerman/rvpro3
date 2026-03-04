package service

import (
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

type ServiceWarningMixin struct {
	Warning string `json:"Warning"`
}

func (u *ServiceWarningMixin) SetWarning(s utils.IRunnableService, msg string) {
	u.Warning = msg
	log.Warn().Str("Service", s.GetServiceName()).Msg(msg)
}
