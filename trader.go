package investind

import m "investindicator/internal/model"

type Trader interface {
	Buy(category m.Category, code string, qty uint) error
}
