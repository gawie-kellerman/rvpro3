package pvr

import (
	"bytes"
	"sync"
	"time"
)

var captureCache = sync.Pool{
	New: func() interface{} {
		return &CaptureMJPegCacheItem{
			Buffer: bytes.NewBuffer(make([]byte, 0, 1024*32)),
		}
	},
}

type CaptureMJPegCacheItem struct {
	Time   time.Time
	Buffer *bytes.Buffer
}

type CaptureMJPegCache struct {
	Cache    []*CaptureMJPegCacheItem
	Capacity int
}

func (c *CaptureMJPegCache) Init(capacity int) {
	c.Capacity = capacity
	c.Cache = make([]*CaptureMJPegCacheItem, c.Capacity+1)
}

func (c *CaptureMJPegCache) Push(buffer *bytes.Buffer) {
	item := captureCache.Get().(*CaptureMJPegCacheItem)
	item.Buffer.Write(buffer.Bytes())

	c.Cache = append(c.Cache, item)

	if len(c.Cache) == cap(c.Cache) {
		c.PopFront()
	}
}

func (c *CaptureMJPegCache) PopFront() {
	if len(c.Cache) > 0 {
		first := c.Cache[0]
		captureCache.Put(first)
		c.Cache = c.Cache[1:]
	}
}

func (c *CaptureMJPegCache) GetFront() *CaptureMJPegCacheItem {
	if len(c.Cache) == 0 {
		return nil
	}

	return c.Cache[0]
}

func (c *CaptureMJPegCache) Depth() int {
	return len(c.Cache)
}

func (c *CaptureMJPegCache) Clear() {
	for len(c.Cache) > 0 {
		c.PopFront()
	}
}
