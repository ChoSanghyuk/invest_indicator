package event

import "time"

type assetMsg struct {
	assetId uint
	isSell  bool
}

type assetMsgSentInfo struct {
	isMsgSent bool
	price     float64
	sentTime  time.Time
}

var assetMsgCache map[assetMsg]*assetMsgSentInfo
var portMsgCache map[bool]time.Time
var dailyCache int

func init() {
	assetMsgCache = make(map[assetMsg]*assetMsgSentInfo)
	portMsgCache = make(map[bool]time.Time)
}

func hasMsgCache(assetId uint, isSell bool, price float64) bool {

	cache := assetMsgCache[assetMsg{
		assetId: assetId,
		isSell:  isSell,
	}]

	if cache == nil { // || cache.sentTime.IsZero() sentTime이 미존재할 경우는 없음
		return false
	}

	diff := time.Since(cache.sentTime)
	if diff <= 6*time.Hour { // 유효한 캐시
		return cache.price == price && cache.isMsgSent
	}

	return false
}

func setMsgCache(assetId uint, isSell bool, price float64) {

	k := assetMsg{
		assetId: assetId,
		isSell:  isSell,
	}

	cache := assetMsgCache[k]

	if cache == nil {
		assetMsgCache[k] = &assetMsgSentInfo{
			isMsgSent: true,
			price:     price,
			sentTime:  time.Now(),
		}
	} else {
		cache.price = price
		cache.sentTime = time.Now()
	}

}

func hasPortCache(isSell bool) bool {

	sendTime := portMsgCache[isSell]

	if (sendTime == time.Time{} || sendTime.Before(time.Now().Add(-2*time.Hour))) { // 보낸 시간이 2시간보다 전이라면
		return false
	} else {
		return true
	}
}

func setPortCache(isSell bool) {
	portMsgCache[isSell] = time.Now()
}

func hasDailyCache() bool {

	if dailyCache == 0 || dailyCache != time.Now().Day() {
		return false
	} else {
		return true
	}
}

func setDailyCache() {
	dailyCache = time.Now().Day()
}
