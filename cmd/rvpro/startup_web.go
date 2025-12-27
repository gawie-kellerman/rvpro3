package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"rvpro3/radarvision.com/internal/config"
	"rvpro3/radarvision.com/internal/config/globalkey"
	"rvpro3/radarvision.com/internal/smartmicro/instrumentation"
	"rvpro3/radarvision.com/utils"
)

func startupWeb() {
	host := config.RVPro.GlobalStr(globalkey.HttpHost)
	if host == "" {
		return
	}

	router := gin.Default()
	router.GET("/general/version", getGeneralVersion)
	router.GET("/metrics/radar", getMetricsRadar)
	router.GET("/metrics/udp", getMetricsUDP)
	err := router.Run(host)

	if err != nil {
		utils.Debug.Panic(err)
	}
}

func getMetricsUDP(context *gin.Context) {
	m := instrumentation.GlobalUDPMetrics
	context.JSON(http.StatusOK, &m)
}

func getMetricsRadar(context *gin.Context) {
	m := instrumentation.GlobalRadarMetrics
	context.JSON(http.StatusOK, &m)
	//bytes, err := json.Marshal(&m)
	//if err != nil {
	//	context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//}
	//context.Data(200, "application/json", bytes)
}

func getGeneralVersion(context *gin.Context) {
	context.String(200, "3.0.0 - Build 125")
}
