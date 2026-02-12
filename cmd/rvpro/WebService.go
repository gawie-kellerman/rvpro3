package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

const webEnabled = "Http.Enabled"
const webHost = "Http.Host"
const WebServiceName = "Web.Service"

type WebService struct {
}

func (w *WebService) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsStr(webHost, "0.0.0.0:8080")
	config.SetSettingAsBool(webEnabled, true)
}

func (w *WebService) SetupAndStart(state *utils.State, config *utils.Settings) {
	if !config.GetSettingAsBool(webEnabled) {
		return
	}

	host := config.GetSettingAsStr(webHost)

	router := gin.Default()
	router.GET("/general/version", w.getGeneralVersion)
	router.GET("/metrics/section", w.getMetricsSection)
	router.GET("/metrics/sections", w.getMetricsSections)
	router.GET("/state/keys", w.getStateKeys)
	router.GET("/state/key", w.getStateKey)

	//router.PUT("/executor/radars/stop", putStopRadars)
	//router.PUT("/executor/radars/start", putStartRadars)
	//router.GET("/executor/radars/status", getRadarsStatus)

	go func() {
		if err := router.Run(host); err != nil {
			// Handle the error if the server fails to start
			utils.Debug.Panic(err)
		}
	}()

	log.Info().Msgf("Web service listening on %s", host)

	state.Set(WebServiceName, host)
}

func (w *WebService) GetServiceName() string {
	return WebServiceName
}

func (w *WebService) GetServiceNames() []string {
	return nil
}

func (w *WebService) getGeneralVersion(context *gin.Context) {
	context.String(200, "3.0.0 - Build 125")
}

func (w *WebService) getMetricsSections(context *gin.Context) {
	context.JSON(http.StatusOK, utils.GlobalMetrics.Names())
}

func (w *WebService) getMetricsSection(context *gin.Context) {
	sectionNames := context.QueryArray("sn")
	result := make(map[string]*utils.Metrics)

	for _, sectionName := range sectionNames {
		section := utils.GlobalMetrics.FindOrNil(sectionName)

		if section != nil {
			result[section.Name] = section
		}
	}
	context.JSON(http.StatusOK, result)
}

func (w *WebService) getStateKey(context *gin.Context) {
	id := context.Query("id")
	result := utils.GlobalState.Get(id)
	context.JSON(http.StatusOK, result)
}

func (w *WebService) getStateKeys(context *gin.Context) {
	result := utils.GlobalState.GetKeys()
	context.JSON(http.StatusOK, result)
}
