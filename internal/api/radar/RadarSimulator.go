package radar

import (
	"net"
	"os"

	"github.com/pkg/errors"
	"rvpro3/radarvision.com/utils"
)

type RadarSimulator struct {
	RadarIP4  utils.IP4
	ServerIP4 utils.IP4
	conn      *net.UDPConn
}

func (r *RadarSimulator) connect() (res *net.UDPConn, err error) {
	if r.conn != nil {
		return r.conn, nil
	}

	radar := r.RadarIP4.ToUDPAddr()
	server := r.ServerIP4.ToUDPAddr()
	r.conn, err = net.DialUDP("udp4", &radar, &server)
	return r.conn, err
}

func (r *RadarSimulator) SendBytes(source []byte) error {
	var written int

	cnx, err := r.connect()

	if err != nil {
		return err
	}

	if written, err = cnx.Write(source); err != nil && written != len(source) {
		err = errors.Errorf(
			"Write %d bytes from %s to %s, expected %d",
			written,
			r.RadarIP4,
			r.ServerIP4,
			len(source),
		)
	}
	return err
}

func (r *RadarSimulator) SendStr(source string) error {
	return r.SendBytes([]byte(source))
}

func (r *RadarSimulator) Close() {
	if r.conn != nil {
		_ = r.conn.Close()
		r.conn = nil
	}
}

func (r *RadarSimulator) SendFile(fileName string) error {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return r.SendBytes(bytes[:])
}
