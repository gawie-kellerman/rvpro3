package tcphub

import (
	"rvpro3/radarvision.com/utils"
)

type PacketBuffer struct {
	queue utils.QueueBuffer
}

func NewBuffer(size int) PacketBuffer {
	res := PacketBuffer{}
	res.Init(size)
	return res
}

func (b *PacketBuffer) Init(bufferSize int) {
	b.queue.Init(bufferSize)
}

func (b *PacketBuffer) PushBytes(data []byte) {
	_ = b.queue.PushData(data, true)
	// TODO: Track failure
}

func (b *PacketBuffer) Pop(packet *Packet) bool {
	slice := b.queue.GetDataSlice()

	// Not enough bytes for the header
	if len(slice) < headerSize {
		return false
	}

	parseStart := getPacketDelimiter(slice)
	if parseStart != startDelimiter {
		//TODO: Track failure
		b.queue.Reset()
		return false
	}

	parseSize := int(getPacketSize(slice))

	// Not enough bytes for the data
	if len(slice) < parseSize {
		return false
	}

	if readSize, err := PacketBuilder.Deserialize(packet, slice); err != nil {
		b.queue.Reset()
		//TODO: Track failure
		return false
	} else {
		if readOver := b.queue.PopSize(readSize); readOver != nil {
			//TODO: Track failure - impossible for code to get here unless bug in PacketBuilder.Deserialize
			b.queue.Reset()
			return false
		}
	}
	return true
}

func (b *PacketBuffer) Size() int {
	return b.queue.Size()
}
