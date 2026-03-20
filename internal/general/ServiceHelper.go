package general

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

type serviceHelper struct {
}

var ServiceHelper serviceHelper

//func (serviceHelper) LogServiceAlreadyDeprecated(service utils.IRunnableService) {
//	log.Warn().Str("Service", service.GetServiceName()).Msg("Service already running")
//}

func (serviceHelper) ShouldStart(state *utils.State, settings *utils.Settings, service utils.IRunnableService) bool {
	if state.Has(service.GetServiceName()) {
		log.Warn().Str("Service", service.GetServiceName()).Msg("Service already running")
		return false
	}
	state.Set(service.GetServiceName(), service)
	//service.InitFromSettings(settings)  <-- Removed for radar-startsim
	return true
}

func (serviceHelper) NameWithIP(name string, ip4 utils.IP4) string {
	return fmt.Sprintf("%s-%s", name, ip4)
}
