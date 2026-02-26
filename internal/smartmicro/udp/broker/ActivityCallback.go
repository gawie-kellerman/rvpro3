package broker

import "rvpro3/radarvision.com/utils"

type ActivityCallback func(item *ActivityCallbackItem, radarIP utils.IP4, payload []byte)

type ActivityCallbackItem struct {
	Index        int
	ActivityType int
	Callback     ActivityCallback
	Name         string
}
type ActivityCatalog struct {
	Catalog map[int][]*ActivityCallbackItem
}

func (a *ActivityCatalog) Register(activityType int, name string, callback ActivityCallback) {
	cat, ok := a.Catalog[activityType]
	if !ok {
		cat = make([]*ActivityCallbackItem, 0)
		a.Catalog[activityType] = cat
	}
	item := &ActivityCallbackItem{
		Index:        len(cat),
		ActivityType: activityType,
		Callback:     callback,
		Name:         name,
	}
	cat = append(cat, item)
}

func (a *ActivityCatalog) Call(activityType int, radarIP utils.IP4, payload []byte) {
	cats, ok := a.Catalog[activityType]
	if !ok {
		return
	}

	for _, cat := range cats {
		if cat.Callback != nil {
			cat.Callback(cat, radarIP, payload)
		}
	}
}
