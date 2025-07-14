package model

type MarketLevel uint

const (
	MAJOR_BEAR MarketLevel = iota + 1
	BEAR
	VOLATILIY
	BULL
	MAJOR_BULL
)

var marketLevelList = []string{"MAJOR_BEAR", "BEAR", "VOLATILIY", "BULL", "MAJOR_BULL"}

func (c MarketLevel) MaxVolatileAssetRate() float64 {
	switch c {
	case MAJOR_BEAR:
		return 0.15
	case BEAR:
		return 0.2
	case VOLATILIY:
		return 0.25
	case BULL:
		return 0.3
	case MAJOR_BULL:
		return 0.4
	}
	return 0
}

func (c MarketLevel) MinVolatileAssetRate() float64 {
	switch c {
	case MAJOR_BEAR:
		return 0.1
	case BEAR:
		return 0.15
	case VOLATILIY:
		return 0.2
	case BULL:
		return 0.25
	case MAJOR_BULL:
		return 0.3
	}
	return 0
}

func (c MarketLevel) String() string {
	return marketLevelList[c-1]
}
