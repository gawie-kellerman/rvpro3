package triggerpipeline

import (
	"sort"

	"rvpro3/radarvision.com/utils"
)

const TriggerPipelineStateName = "Pipeline"

type TriggerPipeline struct {
	Item []ITriggerPipelineItem
}

func (t *TriggerPipeline) AddItem(item ITriggerPipelineItem) ITriggerPipelineItem {
	if result := t.Find(item.GetName(), item.GetRadarIP()); result != nil {
		return result
	}

	t.Item = append(t.Item, item)
	t.Sort()

	return item
}

// Sort is order based only
// Consider sorting by radar ip first
func (t *TriggerPipeline) Sort() {
	sort.SliceStable(t.Item, func(i, j int) bool {
		return t.Item[i].GetOrder() < t.Item[j].GetOrder()
	})
}

func (t *TriggerPipeline) Find(name string, ip utils.IP4) ITriggerPipelineItem {
	for _, item := range t.Item {
		if item.GetName() == name && item.GetRadarIP().ToU32() == ip.ToU32() {
			return item
		}
	}
	return nil
}

func (t *TriggerPipeline) Execute(display ITriggerDisplay) (uint64, uint64) {
	hi := uint64(0)
	lo := uint64(0)

	for _, item := range t.Item {
		item.UpdateDisplay(display)
		hi, lo = item.Execute(hi, lo)
	}

	return hi, lo
}
