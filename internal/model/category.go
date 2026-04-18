package model

import (
	"errors"
	"slices"
)

type Category uint

const (
	Won Category = iota + 1
	Dollar
	Gold
	ShortTermBond
	DomesticETF
	DomesticStock
	DomesticCoin
	ForeignStock
	ForeignETF
	Leverage
	ForeignCoin
	DomesticGoldETF
)

var categoryList = []string{"현금", "달러", "금", "단기채권", "국내ETF", "국내주식", "국내코인", "해외주식", "해외ETF", "레버리지", "해외코인", "국내금ETF"}
var stableList = []Category{Won, Dollar, Gold, ShortTermBond, DomesticGoldETF}

func (c Category) String() string {
	if c == 0 || int(c) >= len(categoryList) {
		return ""
	}
	return categoryList[c-1]
}

func ToCategory(s string) (Category, error) {

	for i, c := range categoryList {
		if s == c {
			return Category(i + 1), nil
		}
	}
	return 0, errors.New("존재하지 않는 카테고리 번호. 입력 값 :" + s)
}

func (c Category) IsStable() bool {
	if slices.Contains(stableList, c) {
		return true
	} else {
		return false
	}
}

func IsValidCategory(c string) bool {
	for _, category := range categoryList {
		if c == category {
			return true
		}
	}
	return false
}

func CategoryLength() uint64 {
	return uint64(len(categoryList))
}

func CategoryList() []string {
	return categoryList
}

// GetMajorCategory returns the major category classification
func (c Category) GetMajorCategory() string {
	switch c {
	case Won, Dollar, Gold:
		return "안정자산"
	default:
		return "변동자산"
	}
}

// GetMiddleCategory returns the middle category classification
func (c Category) GetMiddleCategory() string {
	switch c {
	case Won, Dollar, ShortTermBond:
		return "현금"
	case Gold, DomesticGoldETF:
		return "금"
	case DomesticETF, DomesticStock, ForeignStock, ForeignETF, Leverage:
		return "ETF/주식"
	case DomesticCoin, ForeignCoin:
		return "코인"
	default:
		return "기타"
	}
}

// GetSmallCategory returns the small category classification
func (c Category) GetSmallCategory() string {
	return c.String()
}
