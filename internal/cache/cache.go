package cache

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
var portMsgCache map[uint]time.Time

// var dailyCache int

func init() {
	assetMsgCache = make(map[assetMsg]*assetMsgSentInfo)
	portMsgCache = make(map[uint]time.Time)
}

func HasMsgCache(assetId uint, isSell bool, price float64) bool {

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

func SetMsgCache(assetId uint, isSell bool, price float64) {

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

func HasPortCache(id uint) bool {

	sendTime := portMsgCache[id]

	if (sendTime == time.Time{} || sendTime.Before(time.Now().Add(-2*time.Hour))) { // 보낸 시간이 2시간보다 전이라면
		return false
	} else {
		return true
	}
}

func SetPortCache(id uint) {
	portMsgCache[id] = time.Now()
}
