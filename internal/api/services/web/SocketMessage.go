package web

import "encoding/json"

const SocketTime uint64 = 1

type SocketPayload struct {
	Subscription uint64
	Payload      []byte
}

type SocketMessage struct {
	Data map[string]interface{}
}

func (m *SocketMessage) SetType(msgType string) {
	m.Data["Type"] = msgType
}

func (m *SocketMessage) GetType(defValue string) string {
	res, ok := m.Data["Type"]
	if !ok {
		return defValue
	}
	return res.(string)
}

func (m *SocketMessage) Set(key string, value string) {
	m.Data[key] = value
}

func (m *SocketMessage) SetInt(key string, value int) {
	m.Data[key] = value
}

func (m *SocketMessage) ToPayload(subscription uint64) *SocketPayload {
	data, _ := json.Marshal(m.Data)

	return &SocketPayload{
		Subscription: subscription,
		Payload:      data,
	}
}

func (m *SocketMessage) LoadBytes(bufData []byte) error {
	m.Data = make(map[string]interface{})
	return json.Unmarshal(bufData, &m.Data)
}

func (m *SocketMessage) Init() {
	m.Data = make(map[string]interface{})
}
