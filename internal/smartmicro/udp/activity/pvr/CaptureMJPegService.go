package pvr

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
	"rvpro3/radarvision.com/utils"
)

const CaptureMJPegServiceName = "CaptureMJPegService"
const captureMJPegEnabled = "capture.mjpeg.enabled"
const captureMJPegUrl = "capture.mjpeg.url"
const captureMJPegConnectTimeout = "capture.mjpeg.connect.timeout"
const captureMJPegErrorDuration = "capture.mjpeg.error.duration"
const captureMJPegErrorCountLog = "capture.mjpeg.error.count.log"

// CaptureMJPegService settings should not come from config file
// but rather from the channel configuration
type CaptureMJPegService struct {
	StreamURL            string
	StreamConnectTimeout int
	Enabled              bool
	Terminate            bool
	Terminated           bool
	ErrorDuration        time.Duration
	ErrorCountLog        int
	transport            http.Transport
	Metrics              CaptureMJPegServiceMetrics
	OnFrameCallback      func(service any, now time.Time, buffer *bytes.Buffer)
	OnErrorCallback      func(service any, now time.Time, err error)
}

type CaptureMJPegServiceMetrics struct {
	StreamReadBytes *utils.Metric
	StreamReadCount *utils.Metric
	ErrorCount      *utils.Metric
	utils.MetricsInitMixin
}

func (c *CaptureMJPegService) SetupDefaults(settings *utils.Settings) {
	settings.SetSettingAsInt(captureMJPegConnectTimeout, 2000)
	settings.SetSettingAsInt(captureMJPegErrorDuration, 1000)
	settings.SetSettingAsInt(captureMJPegErrorCountLog, 60)
}

func (c *CaptureMJPegService) SetupAndStart(state *utils.State, settings *utils.Settings) {
	state.Set(CaptureMJPegServiceName, c)
	c.Metrics.InitMetrics(CaptureMJPegServiceName, &c.Metrics)

	c.Terminate = false
	c.Terminated = false

	if !c.Enabled {
		return
	}

	c.transport = http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c.ErrorDuration = settings.GetSettingAsDuration(captureMJPegErrorDuration, 1000) * time.Millisecond
	c.ErrorCountLog = settings.GetSettingAsIntDef(captureMJPegErrorCountLog, 60)
	c.StreamConnectTimeout = settings.GetSettingAsInt(captureMJPegConnectTimeout)

	go c.run()
}

func (c *CaptureMJPegService) GetServiceName() string {
	return CaptureMJPegServiceName
}

func (c *CaptureMJPegService) GetServiceNames() []string {
	return nil
}

func (c *CaptureMJPegService) run() {
	var err error
	var lastErr error
	var errCount int

	for !c.Terminate {
		err = c.stream()

		if err != nil {
			now := time.Now()
			errCount++
			if !errors.Is(err, lastErr) || errCount > c.ErrorCountLog {
				log.Err(err).Msgf("CaptureMJPegService, Count: %d", errCount)
				errCount = 0
			}
			c.Metrics.ErrorCount.IncAt(1, now)

			if c.OnErrorCallback != nil {
				c.OnErrorCallback(c.Metrics, now, err)
			}

			time.Sleep(c.ErrorDuration)
		}
	}

	c.Terminated = true
}

func (c *CaptureMJPegService) stream() (err error) {
	var resp *http.Response
	client := &http.Client{
		Transport: &c.transport,
		Timeout:   time.Duration(0) * time.Millisecond,
	}

	resp, err = client.Get(c.StreamURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the multipart stream
	var mediaType string
	var params map[string]string
	mediaType, params, err = mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/x-mixed-replace" {
		return errors.New("not multipart/x-mixed-replace")
	}
	boundary := params["boundary"]
	reader := multipart.NewReader(resp.Body, boundary)

	buffer := bytes.NewBuffer(nil)
	buffer.Grow(16 * 1024)

	for !c.Terminated {
		now := time.Now()

		var part *multipart.Part
		if part, err = reader.NextPart(); err != nil {
			if part != nil {
				_ = part.Close()
			}
			return err
		}

		var byteCount int64
		if byteCount, err = io.Copy(buffer, part); err != nil {
			_ = part.Close()
			return err
		}

		c.Metrics.StreamReadBytes.IncAt(byteCount, now)
		c.Metrics.StreamReadCount.IncAt(1, now)
		if c.OnFrameCallback != nil {
			c.OnFrameCallback(c, now, buffer)
		}

		_ = part.Close()
	}
	return nil
}
