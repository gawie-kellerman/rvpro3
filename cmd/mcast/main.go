package main

import (
	"net"

	"github.com/rs/zerolog/log"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "239.144.0.0:60000")
	if err != nil {
		log.Err(err).Msg("Failed to resolve UDP address")
		return
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Err(err).Msg("Failed to listen multicast UDP")
		return
	}
	defer conn.Close()

	_ = conn.SetReadBuffer(4096)

	for {
		buffer := make([]byte, 4096)
		noBytes, srcAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Err(err).Msg("Failed to read from UDP")
			return
		}

		log.Info().Msgf("Read %d bytes from %s", noBytes, srcAddr)
	}
}
