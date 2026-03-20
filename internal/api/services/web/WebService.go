package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/general"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/utils"
)

const webEnabled = "feature.http.service.enabled"
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
	Sockets             *SocketService
	Enabled             bool
	Host                string
	SocketEnabled       bool
	SocketPingEvery     utils.Milliseconds
	SocketPongEvery     utils.Milliseconds
	SocketWriteDeadline utils.Milliseconds
	SocketMaxReadSize   int
	SocketMaxWriteSize  int
}

func (w *WebService) InitFromSettings(settings *utils.Settings) {
	w.Enabled = settings.Basic.GetBool(webEnabled, true)
	w.Host = settings.Basic.Get(webHost, "0.0.0.0:8080")
	w.SocketEnabled = settings.Basic.GetBool(socketsEnabled, true)

	w.SocketPingEvery = settings.Basic.GetMilliseconds(socketPingEvery, 30000)
	w.SocketPongEvery = settings.Basic.GetMilliseconds(socketPongEvery, 60000)
	w.SocketWriteDeadline = settings.Basic.GetMilliseconds(socketWriteDeadline, 3000)
	w.SocketMaxReadSize = settings.Basic.GetInt(socketMaxReadSize, 2*utils.Kilobyte)
	w.SocketMaxWriteSize = settings.Basic.GetInt(socketMaxWriteSize, 2*utils.Kilobyte)
}

func (w *WebService) Start(state *utils.State, settings *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, settings, w) {
		return
	}

	if !w.Enabled {
		return
	}

	if w.SocketEnabled {
		w.Sockets = NewSocketService()
		w.Sockets.PingEvery = w.SocketPingEvery
		w.Sockets.PongEvery = w.SocketPongEvery
		w.Sockets.WriteDeadline = w.SocketWriteDeadline
		w.Sockets.MaxReadSize = int64(w.SocketMaxWriteSize)
		upgrader.ReadBufferSize = w.SocketMaxWriteSize
		upgrader.WriteBufferSize = w.SocketMaxWriteSize
		go w.Sockets.Run()
	}

	router := gin.Default()
	router.GET("/general/version", w.getGeneralVersion)
	router.GET("/metrics/section", w.getMetricsSection)
	router.GET("/metrics/sections", w.getMetricsSections)
	router.GET("/state/keys", w.getStateKeys)
	router.GET("/state/key", w.getStateKey)
	router.PUT("/state/set/phase", w.setPhaseState)

	//router.PUT("/executor/radars/stop", putStopRadars)
	//router.PUT("/executor/radars/start", putStartRadars)
	//router.GET("/executor/radars/status", getRadarsStatus)

	// WebSocket
	router.GET("/socket", w.getSocket)

	go func() {
		if err := router.Run(w.Host); err != nil {
			// Handle the error if the server fails to start
			utils.Debug.Panic(err)
		}
	}()

	log.Info().Msgf("Web service listening on %s", w.Host)

	state.Set(WebServiceName, w)
}

func (w *WebService) GetServiceName() string {
	return WebServiceName
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

func (w *WebService) setPhaseState(context *gin.Context) {
	phases := utils.GlobalState.Get(interfaces.PhaseStateName).(interfaces.IPhaseState)
	if phases == nil {
		http.Error(context.Writer, "Phase state not found", http.StatusNotFound)
		return
	}

	var err error
	var r, g, y uint64

	red := context.Query("red")
	green := context.Query("green")
	yellow := context.Query("yellow")

	if r, err = strconv.ParseUint(red, 16, 64); err != nil {
		goto _errorLabel
	}

	if g, err = strconv.ParseUint(green, 16, 64); err != nil {
		goto _errorLabel
	}

	if y, err = strconv.ParseUint(yellow, 16, 64); err != nil {
		goto _errorLabel
	}

	phases.SetRYG("rest", utils.Uint64(r), utils.Uint64(y), utils.Uint64(g))
	context.JSON(http.StatusOK, phases)
	return

_errorLabel:
	context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	return
}
