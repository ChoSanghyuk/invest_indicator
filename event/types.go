package event

type Event struct {
	Id          int
	Title       string
	Description string
	Active      bool
	Event       func(c chan<- string)
}

func (e *EventHandler) registerEvents() {
	e.events = []*Event{
		{
			Id:          1,
			Title:       "Asset 조회",
			Description: "Asset 가격들을 조회 후 우선 매수/매도 대상 Asset으로 정렬 후 반환",
			Active:      true,
			Event:       e.AssetEvent,
		},
	}
}
