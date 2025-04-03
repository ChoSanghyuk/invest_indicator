package event

type EnrolledEvent struct {
	Id          uint
	Title       string
	Description string
	IsActive    bool
	schedule    string
	Event       func()
}

// todo. 얘네를 언제 실행하고,어떤 주기로 실행할지
func (e *EventHandler) registerEvents() {
	e.enrolledEvents = []*EnrolledEvent{
		{
			Id:          1,
			Title:       "매수 Asset 추천",
			Description: "Asset 가격들을 조회 후 우선 매수 대상 Asset으로 정렬 후 반환",
			schedule:    "0 0 8 * * 1-5",
			Event:       e.AssetRecommendEvent,
		},
		{
			Id:          2,
			Title:       "금 김치 프리미엄",
			Description: "금 가격의 한국 시세와 달러 시세의 차이 확인.\n10% 초과 시 알림. 15% 초과 시, 매도 권자 알림.\n09:00~16:00 10분 주기",
			schedule:    "0 */10 9-16 * * 1-5",
			Event:       e.goldKimchiPremium,
		},
	}

	for _, event := range e.enrolledEvents {
		event.IsActive = e.stg.RetreiveEventIsActive(uint(event.Id))
	}
}
