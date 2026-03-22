package stream

type StandardTick struct {
	Ask       float64     `json:"ask,omitempty"`
	AskSize   float64     `json:"askSize,omitempty"`
	Bid       float64     `json:"bid,omitempty"`
	BidSize   float64     `json:"bidSize,omitempty"`
	LastPrice float64     `json:"lastPrice,omitempty"`
	LastSize  float64     `json:"lastSize,omitempty"`
	Amount    float64     `json:"amount,omitempty"`
	Close     float64     `json:"close,omitempty"`
	Open      float64     `json:"open,omitempty"`
	High      float64     `json:"high,omitempty"`
	Low       float64     `json:"low,omitempty"`
	Count     int         `json:"count,omitempty"`
	Vol       float64     `json:"vol,omitempty"`
	Bids      [][]float64 `json:"bids,omitempty"`
	Asks      [][]float64 `json:"asks,omitempty"`
	Trades    []TradeItem `json:"data,omitempty"`
	Ts        int64       `json:"ts"`
}

type TradeItem struct {
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
	Direction string  `json:"direction"`
}
