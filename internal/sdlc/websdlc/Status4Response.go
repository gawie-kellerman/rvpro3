package websdlc

type Status4Response struct {
	Milliseconds int64  `xml:"ms"`
	Uptime       string `xml:"uptime"`
	Version      string `xml:"version"`
	DeviceId     string `xml:"deviceid"`
	DeviceSerial string `xml:"deviceserial"`
	CallStatus   string `xml:"callstatus"`
	Call         string `xml:"CALL"`
	Red          string `xml:"RED"`
	Green        string `xml:"GRN"`
	Yellow       string `xml:"YEL"`
	Led1         string `xml:"led1"`
	Led2         string `xml:"led2"`
	Led3         string `xml:"led3"`
	Led4         string `xml:"led4"`
}
