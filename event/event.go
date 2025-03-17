package event

import (
	"cmp"
	"fmt"
	m "invest/model"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/robfig/cron"
)

type EventHandler struct {
	stg    Storage
	rt     RtPoller
	dp     DailyPoller
	ch     chan<- string
	events []*Event
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

// pricemap 로컬 변수로 두고, asset event랑 포트폴 event 분리시키기
const (
	AssetSpec  = "0 */15 9-23 * * 1-5"
	RcmdSpec   = "0 0 8 * * 1-5"
	CoinSpec   = "0 */15 8-23 * * 0,6"
	EstateSpec = "0 */15 9-17 * * 1-5"
	IndexSpec  = "0 0 8 * * 1-5"
	EmaSpec    = "0 0 8 * * 2-6" // 화~토
)

const portfolioMsgForm string = "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.2f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.1f)\n\n"

func (e EventHandler) Run() { // todo. 주석해제 필요
	c := cron.New()
	c.AddFunc(AssetSpec, e.AssetEvent)
	c.AddFunc(RcmdSpec, e.AssetRecommendEvent)
	c.AddFunc(CoinSpec, e.CoinEvent)
	c.AddFunc(EstateSpec, e.RealEstateEvent)
	c.AddFunc(IndexSpec, e.IndexEvent)
	c.AddFunc(EmaSpec, e.EmaUpdateEvent)
	c.Start()
}

func (e EventHandler) Events() []*Event {
	return e.events
}

func (e EventHandler) StatusChange(id int, active bool) error {

	done := false
	for _, ev := range e.events {
		if ev.Id == id {
			ev.Active = active
			done = true
			break
		}
	}
	if !done {
		return fmt.Errorf("미존재 Id : %d", id)
	}

	return nil
}

func (e EventHandler) Launch(id int) error {

	for _, ev := range e.events {
		if ev.Id == id {
			if ev.Active {
				ev.Event()
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
	e.assetUpdate(priceMap, ivsmLi)

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
		if a.Category == m.DomesticCoin { // 코인에 대해서만 수행
			msg, err := e.buySellMsg(a.ID, priceMap)
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

func (e EventHandler) AssetRecommendEvent() {

	pm := make(map[uint]float64)
	ivsmLi := make([]m.InvestSummary, 0)
	e.assetUpdate(pm, ivsmLi)

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
		if a.Category == m.Won || a.Category == m.Dollar {
			continue
		}
		cp, err := e.dp.ClosingPrice(asset.Category, asset.Code)
		if err != nil {
			e.ch <- fmt.Sprintf("[EmaUpdateEvent] ClosingPrice 시, 에러 발생. %s", err)
			continue
		}
		e.stg.SaveEmaHist(a.ID, cp)
	}
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

func (e EventHandler) assetUpdate(priceMap map[uint]float64, ivsmLi []m.InvestSummary) {
	// 등록 자산 목록 조회
	assetList, err := e.stg.RetrieveAssetList()
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] RetrieveAssetList 시, 에러 발생. %s", err)
		return
	}
	// priceMap := make(map[uint]float64) // assetId => price

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
	ivsmLi, err = e.stg.RetreiveFundsSummaryOrderByFundId()
	if err != nil {
		e.ch <- fmt.Sprintf("[AssetEvent] RetreiveFundsSummaryOrderByFundId 시, 에러 발생. %s", err)
		return
	}
	if len(ivsmLi) == 0 {
		return
	}

	// 자금별/종목별 현재 총액 갱신
	err = e.updateFundSummarys(ivsmLi, priceMap)
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
		e.stg.UpdateAssetInfo(assetId, "", 0, "", "", pp, 0, 0, 0)
	} else if a.Bottom > pp {
		e.stg.UpdateAssetInfo(assetId, "", 0, "", "", 0, pp, 0, 0)
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
		if hasPortCache(k) && !(r > marketLevel.MaxVolatileAssetRate()) && !(r < marketLevel.MinVolatileAssetRate()) {
			continue
		}
		setPortCache(k)

		os := make([]priority, 0, len(ivsmLi)) // ordered slice

		if r > marketLevel.MaxVolatileAssetRate() { // 매도 메시지

			err = e.loadOrderSlice(&os, pm)
			if err != nil {
				return "", err
			}
			// for _, ivsm := range ivsmLi {
			// 	if ivsm.FundID == k {
			// 		a := &ivsm.Asset
			// 		if a.Category == m.Won || a.Category == m.Dollar {
			// 			continue
			// 		}
			// 		pp := pm[a.ID]
			// 		ap, err := e.stg.RetreiveLatestEma(a.ID)
			// 		if err != nil {
			// 			return "", fmt.Errorf("RetreiveLatestEma, 에러 발생. ID: %d. %w", a.ID, err)
			// 		}
			// 		hp := a.Top
			// 		os = append(os, priority{
			// 			asset: a,
			// 			ap:    ap,
			// 			pp:    pp,
			// 			hp:    hp,
			// 			score: 0.6*((pp-ap)/pp) + 0.4*((pp-hp)/pp),
			// 		})
			// 	}
			// }

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
		} else { // 매수 메시지
			err = e.loadOrderSlice(&os, pm)
			if err != nil {
				return "", err
			}
			// li, err := e.stg.RetrieveTotalAssets()
			// if err != nil {
			// 	return "", fmt.Errorf("RetrieveTotalAssets, 에러 발생. %w", err)
			// }
			// // 매수 시기에는 전체 List 조회. Todo. 여러 자금에 대해서 공통적으로 반복 수행하게 될 수 있음.
			// for _, a := range li {
			// 	if a.Category == m.Won || a.Category == m.Dollar {
			// 		continue
			// 	}
			// 	pp := pm[a.ID]
			// 	ap, err := e.stg.RetreiveLatestEma(a.ID)
			// 	if err != nil {
			// 		return "", fmt.Errorf("RetreiveLatestEma, 에러 발생. ID: %d. %w", a.ID, err)
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

			sb.WriteString(fmt.Sprintf(portfolioMsgForm, // "자금 %d 변동 자산 비중 %s.\n  변동 자산 비율 : %.2f.\n  (%.2f/%.2f)\n  현재 시장 단계 : %s(%.1f)\n\n"
				k,
				"부족",
				r,
				volatile[k],
				volatile[k]+stable[k],
				marketLevel.String(),
				marketLevel.MaxVolatileAssetRate()),
			)

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
		ap, err := e.stg.RetreiveLatestEma(a.ID)
		if err != nil {
			return fmt.Errorf("RetreiveLatestEma, 에러 발생. ID: %d. %w", a.ID, err)
		}
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
