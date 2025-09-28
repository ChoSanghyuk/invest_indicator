package investind

import "github.com/robfig/cron"

const (
	AssetSpec  = "0 */15 9-23 * * 1-5"
	CoinSpec   = "0 */15 8-23 * * 0,6"
	EstateSpec = "0 */15 9-17 * * 1-5"
	DailySpec  = "0 0 7 * * 1-5"
	// IndexSpec  = "0 0 7 * * 1-5"
	// EmaSpec    = "0 0 7 * * 2-6" // 화 - 토
)

const portfolioMsgForm string = "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.3f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.2f)\n\n"

func (e InvestIndicator) Run() {
	e.lg.Info().Msg("Starting EventHandler Run")
	c := cron.New()
	c.AddFunc(AssetSpec, e.runAssetEvent)
	c.AddFunc(CoinSpec, e.runCoinEvent)
	// c.AddFunc(IndexSpec, e.IndexEvent)
	// c.AddFunc(EmaSpec, e.EmaUpdateEvent)
	c.AddFunc(DailySpec, func() {
		e.runIndexEvent()
		e.runEmaUpdateEvent()
		e.runHighYieldSpreadEvent()
		e.runAssetRecommendEvent(false)
		e.runFindNewSP500Event()
	})

	// c.AddFunc(EstateSpec, e.RealEstateEvent)

	for _, enrolled := range e.enrolledEvents {
		if enrolled.schedule == "" {
			continue
		}
		c.AddFunc(enrolled.schedule, func() {
			if enrolled.IsActive {
				enrolled.Event(false)
			}
		})
	}

	c.Start()
	e.lg.Info().Msg("EventHandler Run completed")
}

type EnrolledEvent struct {
	Id          uint
	Title       string
	Description string
	IsActive    bool
	schedule    string
	Event       func(WayOfLaunch)
}

type WayOfLaunch bool

const (
	Manual WayOfLaunch = true
	Auto   WayOfLaunch = false
)

func (e *InvestIndicator) registerEvents() {
	e.enrolledEvents = []*EnrolledEvent{
		{
			Id:          1,
			Title:       "매수 Asset 추천",
			Description: "우선 매수 대상 Asset으로 정렬 후 반환",
			schedule:    "", // "0 0 7 * * 1-5",
			Event:       e.runAssetRecommendEvent,
		},
		{
			Id:          2,
			Title:       "금 김치 프리미엄",
			Description: "금 가격의 한국 시세와 달러 시세의 차이 확인.\n5% 초과 시 알림. 10% 초과 시, 매도 권장 알림.\n오후 3시 실행",
			schedule:    "0 0 15 * * 1-5",
			Event:       e.runGoldKimchiPremium,
		},
		{
			Id:          3,
			Title:       "코인 김치 프리미엄",
			Description: "코인 김치 프리미엄 확인.\n매일 오전 8시~오후 12시 15분 주기로 실행",
			schedule:    "0 */15 8-23 * * 0-6",
			Event:       e.runCoinKimchiPremiumEvent,
		},
		{
			Id:          4,
			Title:       "AVAX DEX 관리",
			Description: "AVAX DEX 관리 행동 지시.\n매일 오전 8시~오후 12시 1분 주기로 실행",
			schedule:    "0 */1 8-23 * * 0-6",
			Event:       e.runAvaxDexEvent,
		},
		{
			Id:          5,
			Title:       "신규 거래소 에어드랍 이벤트",
			Description: "AVAX DEX 관리 행동 지시.\n매일 오전 8시~오후 12시 10분 주기로 실행",
			schedule:    "0 */10 8-23 * * 0-6",
			Event:       e.runNewlyOpenedAirdropEvent,
		},
		// 보류
		// {
		// 	Id:          5,
		// 	Title:       "정기 분할 매수",
		// 	Description: "AVAX DEX 관리 행동 지시.\n매월 첫째주 월요일 저녁 11시",
		// 	schedule:    "0 0 23 1-7 * 1",
		// 	Event:       e.runBuySP500Event,
		// },
	}

	for _, event := range e.enrolledEvents {
		event.IsActive = e.stg.RetreiveEventIsActive(uint(event.Id))
	}
}
