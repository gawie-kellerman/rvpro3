package sdlc

import (
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

var errWriteMessageDiscarded = errors.New("write message discarded")

type SDLCService struct {
	Serial            SerialConnection
	doneChan          chan bool
	writeChannel      chan []byte
	readBuffer        [1024]byte
	backingBuffer     [2048]byte
	serialBuffer      utils.SerialBuffer
	terminate         bool
	terminateRefCount atomic.Int32
	WritePool         *SDLCWritePool
	OnError           func(*SDLCService, error)
	OnTerminate       func(*SDLCService)
	OnPopMessage      func(*SDLCService, []byte)
	OnWrite           func(*SDLCService, []byte)
}

func (s *SDLCService) Init() {
	s.WritePool = NewSDLCWritePool()
	s.serialBuffer.Buffer = s.backingBuffer[:]
	s.serialBuffer.StartDelim = 0x02
	s.serialBuffer.EndDelim = 0x03
	s.doneChan = make(chan bool)
	s.writeChannel = make(chan []byte, 5)
	s.terminateRefCount.Store(2)
}

func (s *SDLCService) Start() {
	s.Init()
	go s.executeReader()
	go s.executeWriter()
}

func (s *SDLCService) Stop() {
	s.doneChan <- true

	for s.terminateRefCount.Load() > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *SDLCService) executeReader() {
	for !s.terminate {
		if s.Serial.Connect() {
			fmt.Println("reading...")
			readSize := s.Serial.Read(s.readBuffer[:])
			if readSize > 0 {
				if err := s.serialBuffer.Push(s.readBuffer[:readSize]); err != nil {
					if s.OnError != nil {
						s.OnError(s, err)
					}
				} else {
					if readBytes := s.serialBuffer.Pop(); readBytes != nil {
						if s.OnPopMessage != nil {
							s.OnPopMessage(s, readBytes)
						}
					}
				}
			} else {
				// No data was read.
			}
		} else {
			fmt.Println("sleeping for 1 second")
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Println("aborting reader ...")

	s.Serial.Disconnect()
	s.terminateRefCount.Add(-1)

	if s.OnTerminate != nil {
		s.OnTerminate(s)
	}
}

func (s *SDLCService) executeWriter() {
	for {
		select {
		case data := <-s.writeChannel:
			s.writeData(data)

		case <-s.doneChan:
			s.terminate = true
			s.terminateRefCount.Add(-1)
			close(s.writeChannel)
			close(s.doneChan)
			return
		}
	}
}

func (s *SDLCService) writeData(data []byte) {
	s.Serial.Write(data)
	s.WritePool.Release(data)
}

func (s *SDLCService) Write(data []byte) {
	if !s.terminate {
		if len(s.writeChannel) < cap(s.writeChannel) {
			buffer := s.WritePool.Alloc()
			copy(buffer[0:len(data)], data)
			buffer = buffer[:len(data)]
			s.writeChannel <- buffer
		} else {
			s.WritePool.Release(data)
			log.Err(errWriteMessageDiscarded).Str("msg", hex.EncodeToString(data))
		}
	}
}
