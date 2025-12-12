package hive

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)
import "github.com/gorilla/websocket"

type RVProSocket struct {
	Url             string
	subscriptionBuf [14]byte
	conn            *websocket.Conn
}

func (r *RVProSocket) Init(url string) {
	r.Url = url
}

func (r *RVProSocket) Connect() (err error) {
	r.Disconnect()

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	r.conn, _, err = dialer.Dial(r.Url, nil)
	return err
}

func (r *RVProSocket) Disconnect() {
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RVProSocket) ReadRadarStats() (res []*RVProRadarStat, err error) {
	subscription := r.makeSubscription(2)

	if err = r.conn.WriteMessage(websocket.BinaryMessage, subscription); err != nil {
		return nil, err
	}

	return r.readUntilSubscription()
}

func (r *RVProSocket) makeSubscription(subscription uint64) []byte {
	binary.LittleEndian.PutUint16(r.subscriptionBuf[0:2], 64)
	binary.LittleEndian.PutUint16(r.subscriptionBuf[2:4], 8)
	binary.LittleEndian.PutUint64(r.subscriptionBuf[4:12], subscription)
	r.subscriptionBuf[12] = 3
	r.subscriptionBuf[13] = 2
	return r.subscriptionBuf[:]
}

func (r *RVProSocket) readUntilSubscription() (res []*RVProRadarStat, err error) {
	for n := 0; n < 10; n++ {
		res, err = r._readUntilSubscription()
		if err == nil {
			return res, nil
		}
	}
	return nil, err
}

func (r *RVProSocket) _readUntilSubscription() ([]*RVProRadarStat, error) {
	extractJson := func(buf []byte) ([]*RVProRadarStat, error) {
		jsonStr := string(buf)
		firstIndex := strings.IndexByte(jsonStr, '{')
		lastIndex := strings.LastIndexByte(jsonStr, '}') + 1

		if firstIndex == -1 || lastIndex == -1 {
			return nil, errors.New("invalid json format")
		}

		jsonBuf := jsonStr[firstIndex:lastIndex]
		var jsonMap map[string]any

		if err := json.Unmarshal([]byte(jsonBuf), &jsonMap); err != nil {
			return nil, err
		}

		res := make([]*RVProRadarStat, 0, 4)

		radars, ok := jsonMap["radars"].([]any)
		if !ok {
			return nil, errors.New("invalid json format")
		}

		for _, radar := range radars {
			radarNode := radar.(map[string]any)
			radarStat := &RVProRadarStat{}
			radarStat.Parse(radarNode)

			res = append(res, radarStat)
		}
		return res, nil
	}

	_, message, err := r.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return extractJson(message)
}
