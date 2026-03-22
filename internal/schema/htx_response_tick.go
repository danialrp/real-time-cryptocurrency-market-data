package schema

type BBO struct {
	Ask     float64 `json:"ask"`
	AskSize float64 `json:"askSize"`
	Bid     float64 `json:"bid"`
	BidSize float64 `json:"bidSize"`
}

type Ticker struct {
	Amount    float64 `json:"amount"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Close     float64 `json:"close"`
	Count     int     `json:"count"`
	High      float64 `json:"high"`
	LastPrice float64 `json:"lastPrice"`
	LastSize  float64 `json:"lastSize"`
	Low       float64 `json:"low"`
	Open      float64 `json:"open"`
	Vol       float64 `json:"vol"`
}

type Depth struct {
	Bids [][]float64 `json:"bids"`
	Asks [][]float64 `json:"asks"`
}

type Kline struct {
	Amount float64 `json:"amount"`
	Close  float64 `json:"close"`
	Count  int     `json:"count"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Open   float64 `json:"open"`
	Vol    float64 `json:"vol"`
}

type TradeList struct {
	Data []Trade `json:"data"`
	Ts   int64   `json:"ts,omitempty"`
}

type Trade struct {
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
	Direction string  `json:"direction"`
}
