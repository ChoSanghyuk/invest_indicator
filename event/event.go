package event

import (
	"cmp"
	"fmt"
	m "invest/model"
	"log"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/robfig/cron"
)

type EventHandler struct {
	stg            Storage
	rt             RtPoller
	dp             DailyPoller
	ch             chan<- string
	enrolledEvents []*EnrolledEvent
}

type EventHandlerConfig struct {
	Storage     Storage
	RtPoller    RtPoller
	DailyPoller DailyPoller
	Channel     chan<- string
}

func NewEventHandler(conf EventHandlerConfig) *EventHandler {

	eh := &EventHandler{
		stg: conf.Storage,
		rt:  conf.RtPoller,
		dp:  conf.DailyPoller,
		ch:  conf.Channel,
	}
	eh.registerEvents()
	return eh
}

const (
	AssetSpec = "0 */15 9-23 * * 1-5"
	// RcmdSpec   = "0 0 8 * * 1-5"
	CoinSpec   = "0 */15 8-23 * * 0,6"
	EstateSpec = "0 */15 9-17 * * 1-5"
	IndexSpec  = "0 0 8 * * 1-5"
	EmaSpec    = "0 0 7 * * 2-6" // 화~토
)

const portfolioMsgForm string = "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.2f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.1f)\n\n"

func (e EventHandler) Run() {
	c := cron.New()
	c.AddFunc(AssetSpec, e.AssetEvent)
	c.AddFunc(CoinSpec, e.CoinEvent)
	c.AddFunc(IndexSpec, e.IndexEvent)
	c.AddFunc(EmaSpec, e.EmaUpdateEvent)
	// c.AddFunc(RcmdSpec, e.AssetRecommendEvent)
	// c.AddFunc(EstateSpec, e.RealEstateEvent)

	for _, enrolled := range e.enrolledEvents {
		c.AddFunc(enrolled.schedule, func() {
			if enrolled.IsActive {
				enrolled.Event(false)
			}
		})
	}

	c.Start()
}

func (e EventHandler) Events() []*EnrolledEvent {
	return e.enrolledEvents
}

func (e EventHandler) StatusChange(id uint, active bool) error {

	done := false
	for _, ev := range e.enrolledEvents {
		if ev.Id == id {
			ev.IsActive = active
			e.stg.UpdateEventIsActive(ev.Id, ev.IsActive)
			done = true
			break
		}
	}
	if !done {
		return fmt.Errorf("미존재 Id : %d", id)
	}

	return nil
}

func (e EventHandler) Launch(id uint) error {

	for _, ev := range e.enrolledEvents {
		if ev.Id == id {
			if ev.IsActive {
				ev.Event(true)
				return nil
			} else {
				return fmt.Errorf("비활성화 이벤트 Id: %d", id)
			}
		}
	}

	return nil
}

/*
작업 1. 자산의 현재가와 자산의 매도/매수 기준 비교하여 알림 전송
  - 보유 자산 list
  - 자산 정보
  - 현재가

작업 2. 자금별/종목별 현재 총액 갱신 + 최저가/최고가 갱신
  - investSummary list
  - 현재가
  - 환율
  - 자산 정보

직업 3. 현재 시장 단계에 맞는 변동 자산을 가지고 있는지 확인하여 알림 전송. 대상 시, 우선처분 대상 및 보유 자산 현환 전송
  - 시장 단계
  - 갱신된 investSummary list
*/

func (e EventHandler) AssetEvent() {

	priceMap := make(map[uint]float64)
	ivsmLi := make([]m.InvestSummary, 0)
	e.assetUpdate(priceMap, &ivsmLi)

	// 현재 시장 단계 이하로 변동 자산을 가지고 있는지 확인. (알림 전송)
	msg, err := e.portfolioMsg(ivsmLi, priceMap)
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] portfolioMsg시, 에러 발생. %s", err)
	}
	if msg != "" {
		e.ch <- msg
	}

}

func (e EventHandler) CoinEvent() {

	// 등록 자산 목록 조회
	assetList, err := e.stg.RetrieveAssetList()
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] RetrieveAssetList 시, 에러 발생. %s", err)
		return
	}
	priceMap := make(map[uint]float64)

	// 등록 자산 매수/매도 기준 충족 시, 채널로 메시지 전달
	for _, a := range assetList {
		msg, err := e.buySellMsg(a.ID, priceMap)
		if a.Category == m.DomesticCoin { // 코인에 대해서만 수행
			if err != nil {
				e.ch <- fmt.Sprintf("[AssetEvent] buySellMsg시, 에러 발생. %s", err)
				return
			}
			if msg != "" {
				e.ch <- msg
			}
		}

	}
}

func (e EventHandler) AssetRecommendEvent(isManual bool) {

	pm := make(map[uint]float64)
	ivsmLi := make([]m.InvestSummary, 0)
	e.assetUpdate(pm, &ivsmLi) // memo. Map과 slice는 둘 다 reference type이므로, 함수에 넘긴 후의 변경 사항이 원본에도 반영. 단, slice의 경우 capacity를 넘기면 별도로 구성되어 원본 영향 X

	os := make([]priority, 0, len(ivsmLi))
	err := e.loadOrderSlice(&os, pm)
	if err != nil {
		e.ch <- err.Error()
	}
	// li, err := e.stg.RetrieveTotalAssets()
	// if err != nil {
	// 	e.ch <- fmt.Sprintf("RetrieveTotalAssets, 에러 발생. %s", err.Error())
	// 	return
	// }
	// os := make([]priority, 0, len(li)) // ordered slice
	// // 매수 시기에는 전체 List 조회. Todo. 여러 자금에 대해서 공통적으로 반복 수행하게 될 수 있음.
	// for _, a := range li {
	// 	pp := pm[a.ID]
	// 	ap, err := e.stg.RetreiveLatestEma(a.ID)
	// 	if err != nil {
	// 		e.ch <- fmt.Sprintf("RetreiveLatestEma, 에러 발생. ID: %d. %s", a.ID, err.Error())
	// 		return
	// 	}
	// 	hp := a.Top
	// 	os = append(os, priority{
	// 		asset: &a,
	// 		ap:    ap,
	// 		pp:    pp,
	// 		hp:    hp,
	// 		score: 0.6*((pp-ap)/pp) + 0.4*((pp-hp)/pp),
	// 	})
	// }

	slices.SortFunc(os, func(a, b priority) int {
		return cmp.Compare(a.score, b.score)
	})

	var sb strings.Builder
	for _, p := range os {
		sb.WriteString(fmt.Sprintf("AssetId : %d\n  AssetName : %s\n  PresentPrice : %.2f\n  WeighedAveragePrice : %.2f\n  HighestPrice : %.2f\n\n", p.asset.ID, p.asset.Name, p.pp, p.ap, p.hp))
	}

	e.ch <- sb.String()
}

func (e EventHandler) EmaUpdateEvent() {

	// 등록 자산 목록 조회
	assetList, err := e.stg.RetrieveAssetList()
	if err != nil {
		e.ch <- fmt.Sprintf("[EmaUpdateEvent] RetrieveAssetList 시, 에러 발생. %s", err)
		return
	}

	for _, a := range assetList {
		asset, err := e.stg.RetrieveAsset(a.ID)
		if err != nil {
			e.ch <- fmt.Sprintf("[EmaUpdateEvent] RetrieveAsset 시, 에러 발생. %s", err)
			return
		}
		// EMA 갱신 제외
		if asset.Category == m.Won || asset.Category == m.Dollar {
			continue
		}

		cp, err := e.dp.ClosingPrice(asset.Category, asset.Code)
		if err != nil {
			e.ch <- fmt.Sprintf("[EmaUpdateEvent] ClosingPrice 시, 에러 발생. %s", err)
			continue
		}

		oldEma, err := e.stg.RetreiveLatestEma(asset.ID)
		if err != nil {
			e.ch <- fmt.Sprintf("[EmaUpdateEvent] RetreiveLatestEma 시, 에러 발생. %s", err)
			continue
		}

		newEma := CalEma(oldEma, cp)
		if newEma == nil {
			continue
		}

		err = e.stg.SaveEmaHist(newEma)
		if err != nil {
			e.ch <- fmt.Sprintf("[EmaUpdateEvent] SaveEmaHist 시, 에러 발생. %s", err)
			continue
		}
	}
}

/*
a =  10/N+1
EMAt = a*PRICEt + (1-a)EMAy
*/
func CalEma(oldEma *m.EmaHist, cp float64) (newEma *m.EmaHist) {

	var nDays uint
	var ema float64

	if oldEma.NDays == 0 || oldEma.Ema == 0 { // 0 일 때는 스킵
		return nil
	} else if oldEma.NDays < 200 {
		ema = (cp + oldEma.Ema*float64(nDays)) / float64(nDays+1) // 200일 되기 전까진 exponential 미존재
		nDays = oldEma.NDays + 1
	} else {
		nDays = 200
		a := 2.0 / (float64(nDays) + 1) // 200일 이후 부터는 현재가에 가중치 부여.
		ema = math.Round((a*cp+(1-a)*oldEma.Ema)*100) / 100
	}

	newEma = &m.EmaHist{
		AssetID: oldEma.AssetID,
		Ema:     ema,
		NDays:   nDays,
	}

	return newEma
}

func (e EventHandler) RealEstateEvent() {

	rtn, err := e.rt.RealEstateStatus()
	if err != nil {
		e.ch <- fmt.Sprintf("[RealEstateEvent] 크롤링 시 오류 발생. %s", err.Error())
		return
	}

	if rtn != "예정지구 지정" {
		e.ch <- fmt.Sprintf("연신내 재개발 변동 사항 존재. 예정지구 지정 => %s", rtn)
	} else {
		log.Printf("연신내 변동 사항 없음. 현재 단계: %s", rtn)
	}
}

func (e EventHandler) IndexEvent() {

	// 1. 공포 탐욕 지수
	fgi, err := e.dp.FearGreedIndex()
	if err != nil {
		e.ch <- fmt.Sprintf("공포 탐욕 지수 조회 시 오류 발생. %s", err.Error())
		return
	}
	// 2. Nasdaq 지수 조회
	nasdaq, err := e.dp.Nasdaq()
	if err != nil {
		e.ch <- fmt.Sprintf("Nasdaq Index 조회 시 오류 발생. %s", err.Error())
		return
	}

	// 3. SP 지수 조회
	sp500, err := e.dp.Sp500()
	if err != nil {
		e.ch <- fmt.Sprintf("S&P 500 Index 조회 시 오류 발생. %s", err.Error())
		return
	}

	// 오늘분 저장
	err = e.stg.SaveDailyMarketIndicator(fgi, nasdaq, sp500)
	if err != nil {
		e.ch <- fmt.Sprintf("Nasdaq Index 저장 시 오류 발생. %s", err.Error())
	}

	// 어제꺼 조회
	var former string
	if time.Now().Weekday() == 1 {
		former = time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	} else {
		former = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	di, _, err := e.stg.RetrieveMarketIndicator(former)
	if err != nil {
		e.ch <- fmt.Sprintf("금일 공포 탐욕 지수 : %d\n금일 Nasdaq : %.2f", fgi, nasdaq)
	} else {
		e.ch <- fmt.Sprintf("금일 공포 탐욕 지수 : %d (전일 : %d)\n금일 Nasdaq : %.2f\n   (전일 : %.2f)", fgi, di.FearGreedIndex, nasdaq, di.NasDaq)
	}
}

/**********************************************************************************************************************
*********************************************Inner Function************************************************************
**********************************************************************************************************************/

func (e EventHandler) assetUpdate(priceMap map[uint]float64, ivsmLi *[]m.InvestSummary) {
	// 등록 자산 목록 조회
	assetList, err := e.stg.RetrieveAssetList()
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] RetrieveAssetList 시, 에러 발생. %s", err)
		return
	}

	// 등록 자산 매수/매도 기준 충족 시, 채널로 메시지 전달
	for _, a := range assetList {
		msg, err := e.buySellMsg(a.ID, priceMap)
		if err != nil {
			e.ch <- fmt.Sprintf("[AssetEvent] buySellMsg시, 에러 발생. %s", err)
			return
		}
		if msg != "" {
			e.ch <- msg
		}
	}

	// 자금별 종목 투자 내역 조회
	*ivsmLi, err = e.stg.RetreiveFundsSummaryOrderByFundId()
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] RetreiveFundsSummaryOrderByFundId 시, 에러 발생. %s", err)
		return
	}
	if len(*ivsmLi) == 0 {
		return
	}

	// 자금별/종목별 현재 총액 갱신
	err = e.updateFundSummarys(*ivsmLi, priceMap)
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] updateFundSummary 시, 에러 발생. %s", err)
		return
	}
}

func (e EventHandler) buySellMsg(assetId uint, pm map[uint]float64) (msg string, err error) {

	// 자산 정보 조회
	a, err := e.stg.RetrieveAsset(assetId)
	if err != nil {
		return "", fmt.Errorf("[AssetEvent] RetrieveAsset 시, 에러 발생. %w", err)
	}

	// 자산별 현재 가격 조회
	pp, err := e.rt.PresentPrice(a.Category, a.Code)
	if err != nil {
		return "", fmt.Errorf("[AssetEvent] PresentPrice 시, 에러 발생. %w", err)
	}

	pm[assetId] = pp

	// 자산 매도/매수 기준 비교 및 알림 여부 판단. (알림 전송)
	if a.BuyPrice >= pp && !hasMsgCache(a.ID, false, a.BuyPrice) {
		msg = fmt.Sprintf("BUY %s. ID : %d. LOWER BOUND : %.2f. CURRENT PRICE :%.2f", a.Name, a.ID, a.BuyPrice, pp)
		setMsgCache(a.ID, false, a.BuyPrice)
	} else if a.SellPrice != 0 && a.SellPrice <= pp && e.hasIt(a.ID) && !hasMsgCache(a.ID, true, a.SellPrice) {
		msg = fmt.Sprintf("SELL %s. ID : %d. UPPER BOUND : %.2f. CURRENT PRICE :%.2f", a.Name, a.ID, a.SellPrice, pp)
		setMsgCache(a.ID, true, a.SellPrice)
	}

	// 최고가/최저가 갱신 여부 판단
	if a.Top < pp {
		e.stg.UpdateAssetInfo(m.Asset{
			ID:  assetId,
			Top: pp,
		})
	} else if a.Bottom > pp {
		e.stg.UpdateAssetInfo(m.Asset{
			ID:     assetId,
			Bottom: pp,
		})
	}

	return
}

func (e EventHandler) hasIt(id uint) bool {
	li, err := e.stg.RetreiveFundSummaryByAssetId(id)
	if err != nil {
		return true // db 문제로 오류 발생 시, 우선 보유 가정
	}

	cnt := 0.0
	for _, e := range li {
		cnt += e.Count
	}
	if cnt > 0 {
		return true
	} else {
		return false
	}

}

func (e EventHandler) updateFundSummarys(list []m.InvestSummary, pm map[uint]float64) (err error) {
	for i := range len(list) {
		is := &list[i]
		is.Sum = pm[is.AssetID] * float64(is.Count)

		err = e.stg.UpdateInvestSummarySum(is.FundID, is.AssetID, is.Sum)
		if err != nil {
			return
		}
	}
	return nil
}

type priority struct {
	asset *m.Asset
	ap    float64
	pp    float64
	hp    float64
	score float64
}

func (e EventHandler) portfolioMsg(ivsmLi []m.InvestSummary, pm map[uint]float64) (msg string, err error) {
	// 현재 시장 단계 조회
	market, err := e.stg.RetrieveMarketStatus("")
	if err != nil {
		msg = fmt.Sprintf("[portfolioMsg] RetrieveMarketStatus 시, 에러 발생. %s", err)
		return
	}
	marketLevel := m.MarketLevel(market.Status)

	// 환율까지 계산하여 원화로 변환
	ex := e.dp.ExchageRate()
	if ex == 0 {
		msg = "[ExchageRate] ExchageRate 시 환율 값 0 반환"
		return
	}

	keySet := make(map[uint]bool)
	stable := make(map[uint]float64)
	volatile := make(map[uint]float64)

	for i := range len(ivsmLi) {

		ivsm := &ivsmLi[i]

		keySet[ivsm.FundID] = true

		// 원화 가치로 환산
		var v float64
		if ivsm.Asset.Currency == m.USD.String() {
			v = ivsm.Sum * ex
		} else {
			v = ivsm.Sum
		}

		// 자금 종류별 안전 자산 가치, 변동 자산 가치 총합 계산
		if ivsm.Asset.Category.IsStable() {
			stable[ivsm.FundID] = stable[ivsm.FundID] + v
		} else {
			volatile[ivsm.FundID] = volatile[ivsm.FundID] + v
		}
	}

	var sb strings.Builder
	for k := range keySet {

		if volatile[k]+stable[k] == 0 {
			continue
		}

		r := volatile[k] / (volatile[k] + stable[k])
		if hasPortCache(k) || (r > marketLevel.MinVolatileAssetRate() && r < marketLevel.MaxVolatileAssetRate()) { // 캐시가 있거나, 범주안에 있으면 스킵
			continue
		}
		setPortCache(k)

		os := make([]priority, 0, len(ivsmLi)) // ordered slice
		err = e.loadOrderSlice(&os, pm)
		if err != nil {
			return "", err
		}

		if r > marketLevel.MaxVolatileAssetRate() { // 매도 메시지

			sb.WriteString(fmt.Sprintf(portfolioMsgForm, // "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.2f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.1f)\n\n"
				k,
				"초과",
				r,
				volatile[k],
				volatile[k]+stable[k],
				marketLevel.String(),
				marketLevel.MaxVolatileAssetRate()),
			)

			slices.SortFunc(os, func(a, b priority) int {
				if a.asset.Category.IsStable() == b.asset.Category.IsStable() {
					return cmp.Compare(b.score, a.score) // 큰 게 앞으로
				} else {
					if a.asset.Category.IsStable() {
						return 1
					} else {
						return -1
					}
				}
			})
		} else {
			if r < marketLevel.MinVolatileAssetRate() { // 매수 메시지
				sb.WriteString(fmt.Sprintf(portfolioMsgForm, // "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.2f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.1f)\n\n"
					k,
					"부족",
					r,
					volatile[k],
					volatile[k]+stable[k],
					marketLevel.String(),
					marketLevel.MinVolatileAssetRate()),
				)
			}

			slices.SortFunc(os, func(a, b priority) int {
				return cmp.Compare(a.score, b.score)
			})
		}

		for _, p := range os {
			sb.WriteString(fmt.Sprintf("AssetId : %d\n  AssetName : %s\n  PresentPrice : %.2f\n  WeighedAveragePrice : %.2f\n  HighestPrice : %.2f\n\n", p.asset.ID, p.asset.Name, p.pp, p.ap, p.hp))
		}

	}

	msg = sb.String()
	return msg, nil
}

func (e EventHandler) loadOrderSlice(os *[]priority, pm map[uint]float64) error {
	li, err := e.stg.RetrieveTotalAssets()
	if err != nil {
		return fmt.Errorf("RetrieveTotalAssets, 에러 발생. %w", err)
	}

	// 매수 시기에는 전체 List 조회. Todo. 여러 자금에 대해서 공통적으로 반복 수행하게 될 수 있음.
	for _, a := range li {
		if a.Category == m.Won || a.Category == m.Dollar {
			continue
		}
		pp := pm[a.ID]
		ema, err := e.stg.RetreiveLatestEma(a.ID)
		if err != nil {
			return fmt.Errorf("RetreiveLatestEma, 에러 발생. ID: %d. %w", a.ID, err)
		}
		ap := ema.Ema
		hp := a.Top

		if ap == 0 || hp == 0 {
			continue
		}

		*os = append(*os, priority{
			asset: &a,
			ap:    ap,
			pp:    pp,
			hp:    hp,
			score: 0.6*((pp-ap)/pp) + 0.4*((pp-hp)/pp),
		})
	}
	return nil
}

/*
[판단]
현재가가 고점 및 이평가보다 낮을수록 저평가(조정) => 매수
현재가가 고점 및 이평가보다 높을수록 고평가 => 매도

[수식]
pp - 현재가
ap - 평균가
hp - 최고가

매도매수지수 = 0.6*((pp-ap)/pp) + 0.4*((pp-hp))/pp)
매도매수지수 클수록 매도 우선 순위
매도매수지수 낮을수록 매수 우선순위
*/

func (e EventHandler) goldKimchiPremium(isManual bool) {

	goldAsset, err := e.stg.RetrieveAsset(4)
	if err != nil {
		e.ch <- err.Error() //todo log
	}

	if goldAsset == nil || goldAsset.Category != m.Gold {

		assets, err := e.stg.RetrieveAssetList()
		if err != nil {
			e.ch <- err.Error() //todo log
			return
		}

		for _, a := range assets {
			if a.Category == m.Gold {
				goldAsset = &a
				continue
			}
		}
	}

	kp, err := e.rt.PresentPrice(goldAsset.Category, goldAsset.Code) // kimchi price
	if err != nil {
		//todo log
		e.ch <- fmt.Sprintf("금 한국 가격 조회 시 오류. %s", err.Error())
		return
	}

	dp, err := e.rt.GoldPriceDollar() // dollar price
	if err != nil {
		//todo log
		e.ch <- fmt.Sprintf("금 달러 가격 조회 시 오류. %s", err.Error())
		return
	}
	ex := e.dp.ExchageRate() // exchange rate
	cp := dp * ex            // converted price

	kPrm := 100 * (kp - cp) / cp // k-premium

	if kPrm > 10 {
		e.ch <- fmt.Sprintf("[매도] 금 김치 프리미엄 10프로 초과. 현재 프리미엄: %.2f", kPrm)
	} else if kPrm > 5 {
		e.ch <- fmt.Sprintf("[알림] 금 김치 프리미엄 5프로 초과. 현재 프리미엄: %.2f", kPrm)
	} else if kPrm < -2 {
		e.ch <- fmt.Sprintf("[매수] 금 역 김치 프리미엄 2프로 초과. 현재 프리미엄: %.2f", kPrm)
	}

	if isManual {
		e.ch <- fmt.Sprintf("[알림] 현재 프리미엄: %.2f", kPrm)
	}
}
