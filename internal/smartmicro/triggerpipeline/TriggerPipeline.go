package triggerpipeline

import (
	"sort"
	"time"

	"rvpro3/radarvision.com/utils"
)

const TriggerPipelineStateName = "Pipeline"

type TriggerPipeline struct {
	Item []ITriggerPipelineItem
}

func GetTriggerPipeline() *TriggerPipeline {
	return utils.GlobalState.Get(TriggerPipelineStateName).(*TriggerPipeline)
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

func (t *TriggerPipeline) ListByRadar(radarIP utils.IP4, fill []ITriggerPipelineItem) int {
	target := 0

	for _, item := range t.Item {
		if item.GetRadarIP().ToU32() == radarIP.ToU32() {
			fill[target] = item
			target++
			if target == len(fill) {
				break
			}
		}
	}
	return target
}

func (t *TriggerPipeline) Execute(now time.Time, source utils.Uint128, display ITriggerDisplay) utils.Uint128 {
	res := source

	for _, item := range t.Item {
		res = item.Execute(now, res, display)
	}

	return res
}

func (t *TriggerPipeline) ExecuteList(item []ITriggerPipelineItem) utils.Uint128 {
	now := time.Now()

	res := utils.Uint128{}

	for _, item := range t.Item {
		res = item.Execute(now, res, nil)
	}

	return res
}
