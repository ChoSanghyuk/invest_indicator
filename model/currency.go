package model

import "slices"

type Currency uint

const (
	KRW Currency = iota + 1
	USD
)

var currencyList = []string{"WON", "USD"}

func (c Currency) String() string {
	return currencyList[c-1]
}

func IsCurrency(t string) bool {
	return slices.Contains(currencyList, t)
}
