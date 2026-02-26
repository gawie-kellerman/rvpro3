package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

const webEnabled = "http.enabled"
const webHost = "http.host"
const socketsEnabled = "http.sockets.enabled"
const WebServiceName = "Web.Service"
const socketPingEvery = "http.sockets.ping.every"
const socketPongEvery = "http.sockets.pong.every"
const socketWriteDeadline = "http.sockets.write.deadline"
const socketMaxReadSize = "http.sockets.max.read.size"
const socketMaxWriteSize = "http.sockets.max.write.size"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2000,
	WriteBufferSize: 2000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebService struct {
	Sockets *SocketService
}

func (w *WebService) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsStr(webHost, "0.0.0.0:8080")
	config.SetSettingAsBool(webEnabled, true)
	config.SetSettingAsBool(socketsEnabled, true)
	config.SetSettingAsInt(socketPingEvery, 30000)
	config.SetSettingAsInt(socketPongEvery, 60000)
	config.SetSettingAsInt(socketWriteDeadline, 3000)
	config.SetSettingAsInt(socketMaxReadSize, 2000)
	config.SetSettingAsInt(socketMaxWriteSize, 8000)
}

func (w *WebService) SetupAndStart(state *utils.State, config *utils.Settings) {

	if !config.GetSettingAsBool(webEnabled) {
		return
	}

	if config.GetSettingAsBool(socketsEnabled) {
		w.Sockets = NewSocketService()
		w.Sockets.PingEvery = time.Duration(config.GetSettingAsInt(socketPingEvery)) * time.Millisecond
		w.Sockets.PongEvery = time.Duration(config.GetSettingAsInt(socketPongEvery)) * time.Millisecond
		w.Sockets.WriteDeadline = time.Duration(config.GetSettingAsInt(socketWriteDeadline)) * time.Millisecond
		w.Sockets.MaxReadSize = int64(config.GetSettingAsInt(socketMaxReadSize))
		upgrader.ReadBufferSize = int(w.Sockets.MaxReadSize)
		upgrader.WriteBufferSize = config.GetSettingAsInt(socketMaxWriteSize)
		go w.Sockets.Run()
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

	// WebSocket
	router.GET("/socket", w.getSocket)

	go func() {
		if err := router.Run(host); err != nil {
			// Handle the error if the server fails to start
			utils.Debug.Panic(err)
		}
	}()

	log.Info().Msgf("Web service listening on %s", host)

	state.Set(WebServiceName, w)
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

	regexes := context.QueryArray("regex")
	for _, regex := range regexes {
		utils.GlobalMetrics.MergeRegEx(result, regex)
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

func (w *WebService) getSocket(context *gin.Context) {
	conn, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		log.Err(err).Msg("WebSocket.ServeWebSocket")
		return
	}

	subscribeStr := context.Request.URL.Query().Get("subscribe")
	subscriptions, _ := strconv.Atoi(subscribeStr)

	client := &SocketClient{
		service:       w.Sockets,
		conn:          conn,
		send:          make(chan *SocketPayload, 5),
		Subscriptions: uint64(subscriptions),
	}
	w.Sockets.register <- client

	go client.readSocket()
	go client.writeSocket()
}

func (w *WebService) Broadcast(msg *SocketPayload) {
	if w.Sockets == nil {
		return
	}
	w.Sockets.Broadcast(msg)
}

func (w *WebService) NoSubscriptions(mask uint64) int {
	if w.Sockets == nil {
		return 0
	}

	return w.Sockets.NoSubscriptions(mask)
}

func (w *WebService) IsAnySubscribed(mask uint64) bool {
	if w.Sockets == nil {
		return false
	}
	return w.Sockets.IsAnySubscribed(mask)
}
