package event

type EnrolledEvent struct {
	Id          int
	Title       string
	Description string
	Active      bool
	Event       func()
}

func (e *EventHandler) registerEvents() {
	e.events = []*EnrolledEvent{
		{
			Id:          1,
			Title:       "매수 Asset 추천",
			Description: "Asset 가격들을 조회 후 우선 매수 대상 Asset으로 정렬 후 반환",
			Active:      true,
			Event:       e.AssetRecommendEvent,
		},
	}
}
