package trigger

import (
	"bytes"
	"crypto/tls"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/general"

	"rvpro3/radarvision.com/utils"
)

const MJPegStreamServiceName = "MJPegStream.Service"
const streamMJPegEnabled = "stream.mjpeg.enabled"
const streamMJPegUrl = "stream.mjpeg.url"
const streamMJPegConnectTimeout = "stream.mjpeg.connect.timeout"
const streamMJPegErrorDuration = "stream.mjpeg.error.duration"
const streamMJPegErrorLogRepeat = "stream.mjpeg.error.log.repeat"
const MJPegDefaultIPs = "192.168.11.22:443;192.168.11.23:443;192.168.11.24:443;192.168.11.25:443"

// MJPegStreamService settings should not come from config file
// but rather from the channel configuration
type MJPegStreamService struct {
	CameraIP             utils.IP4                                              `json:"CameraIP"`
	StreamURL            string                                                 `json:"StreamURL"`
	StreamConnectTimeout time.Duration                                          `json:"StreamConnectTimeout"`
	Enabled              bool                                                   `json:"Enabled"`
	Terminate            bool                                                   `json:"Terminate"`
	Terminated           bool                                                   `json:"Terminated"`
	ErrorDuration        time.Duration                                          `json:"ErrorDuration"`
	ErrorCountLog        int                                                    `json:"ErrorCountLog"`
	Metrics              MJPegStreamServiceMetrics                              `json:"-"`
	OnFrameCallback      func(service any, now time.Time, buffer *bytes.Buffer) `json:"-"`
	OnErrorCallback      func(service any, now time.Time, err error)            `json:"-"`
	transport            http.Transport
	ServiceName          string `json:"ServiceName"`
}

type MJPegStreamServiceMetrics struct {
	FrameReadBytes         *utils.Metric
	FrameReadCount         *utils.Metric
	FrameMinDuration       *utils.Metric
	FrameMaxDuration       *utils.Metric
	FrameTotalDuration     *utils.Metric
	ErrorsTotal            *utils.Metric
	ErrorsLogged           *utils.Metric
	ErrorsOfHttpConnect    *utils.Metric
	ErrorsOfContentType    *utils.Metric
	ErrorsOfHttpMultipart  *utils.Metric
	ErrorsOfHttpStreamCopy *utils.Metric
	ConnectCount           *utils.Metric
	ConnectTotalDuration   *utils.Metric
	ConnectMinDuration     *utils.Metric
	ConnectMaxDuration     *utils.Metric
	Callbacks              *utils.Metric
	utils.MetricsInitMixin
}

func (c *MJPegStreamService) InitFromSettings(settings *utils.Settings) {
	camIP := c.CameraIP.String()
	enabled := settings.Basic.GetBool(streamMJPegEnabled, false)
	c.Enabled = settings.Indexed.GetBool(streamMJPegEnabled, camIP, enabled)
	c.StreamConnectTimeout = settings.Indexed.GetDurationMs(streamMJPegConnectTimeout, camIP, 2000)
	c.ErrorDuration = settings.Indexed.GetDurationMs(streamMJPegErrorDuration, camIP, 1000)
	c.ErrorCountLog = settings.Indexed.GetInt(streamMJPegErrorLogRepeat, camIP, 100)
	c.StreamURL = settings.Indexed.Get(streamMJPegUrl, camIP, "")
	_ = settings.Basic.Get("stream.mjpeg.camera.ips", MJPegDefaultIPs)
}

func (c *MJPegStreamService) InitBeforeStart(cameraIP utils.IP4) {
	c.CameraIP = cameraIP
	c.ServiceName = general.ServiceHelper.NameWithIP(MJPegStreamServiceName, cameraIP)
}

func (c *MJPegStreamService) Start(state *utils.State, settings *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, settings, c) {
		return
	}

	c.Metrics.InitMetrics(c.GetServiceName(), &c.Metrics)

	if !c.Enabled || c.StreamURL == "" {
		c.Enabled = false
		return
	}

	c.Terminate = false
	c.Terminated = false

	c.transport = http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	go c.run()
}

func (c *MJPegStreamService) GetServiceName() string {
	return c.ServiceName
}

func (c *MJPegStreamService) run() {
	var err error
	var lastErr error
	var errCount int
	var now time.Time

	for !c.Terminate {
		err = c.stream()

		if err != nil {
			now = time.Now()
			c.Metrics.ErrorsTotal.IncAt(1, now)

			errCount++
			if !errors.Is(err, lastErr) || errCount > c.ErrorCountLog {
				c.Metrics.ErrorsLogged.IncAt(1, now)
				log.Err(err).Msgf("CaptureMJPegService, Count: %d", errCount)
				errCount = 0
			}

			if c.OnErrorCallback != nil {
				c.OnErrorCallback(c.Metrics, now, err)
			}

			time.Sleep(c.ErrorDuration)
		}
	}

	c.Terminated = true
}

func (c *MJPegStreamService) stream() (err error) {
	var resp *http.Response
	client := &http.Client{
		Transport: &c.transport,
		Timeout:   time.Duration(0) * time.Millisecond,
	}

	startConnect := time.Now()

	resp, err = client.Get(c.StreamURL)
	if err != nil {
		c.Metrics.ErrorsOfHttpConnect.Inc(1)
		return err
	}
	defer resp.Body.Close()

	// Parse the multipart stream
	var mediaType string
	var params map[string]string
	mediaType, params, err = mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/x-mixed-replace" {
		c.Metrics.ErrorsOfContentType.Inc(1)
		return errors.New("not multipart/x-mixed-replace")
	}
	boundary := params["boundary"]
	reader := multipart.NewReader(resp.Body, boundary)

	buffer := bytes.NewBuffer(nil)
	buffer.Grow(16 * 1024)

	frameStart := time.Now()
	connectTime := frameStart.Sub(startConnect).Milliseconds()

	c.Metrics.ConnectCount.IncAt(1, frameStart)
	c.Metrics.ConnectTotalDuration.IncAt(connectTime, frameStart)
	c.Metrics.ConnectMinDuration.SetIfLessAt(connectTime, frameStart)
	c.Metrics.ConnectMaxDuration.SetIfMoreAt(connectTime, frameStart)

	for !c.Terminated {
		var part *multipart.Part
		if part, err = reader.NextPart(); err != nil {
			c.Metrics.ErrorsOfHttpMultipart.IncAt(1, frameStart)
			if part != nil {
				_ = part.Close()
			}
			return err
		}

		var byteCount int64
		if byteCount, err = io.Copy(buffer, part); err != nil {
			c.Metrics.ErrorsOfHttpStreamCopy.IncAt(1, frameStart)
			_ = part.Close()
			return err
		}

		// At this point...  everything successful and we have a frame
		frameEnd := time.Now()
		durationMs := frameEnd.Sub(frameStart).Milliseconds()

		c.Metrics.FrameMinDuration.SetIfLessAt(durationMs, frameEnd)
		c.Metrics.FrameMaxDuration.SetIfMoreAt(durationMs, frameEnd)
		c.Metrics.FrameTotalDuration.IncAt(durationMs, frameEnd)
		c.Metrics.FrameReadBytes.IncAt(byteCount, frameEnd)
		c.Metrics.FrameReadCount.IncAt(1, frameEnd)

		if c.OnFrameCallback != nil {
			c.Metrics.Callbacks.IncAt(1, frameStart)
			c.OnFrameCallback(c, frameStart, buffer)
		}

		_ = part.Close()
		frameStart = time.Now()
	}
	return nil
}
