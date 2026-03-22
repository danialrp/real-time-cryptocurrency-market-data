package htx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"irbtc_streamer/internal/model"
)

type rawSymbol struct {
	Symbol        string `json:"symbol"`
	BaseCurrency  string `json:"bc"`
	QuoteCurrency string `json:"qc"`
	State         string `json:"state"`
	TradeEnabled  bool   `json:"te"`
}

func FetchSymbols(ctx context.Context) ([]model.Symbol, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.htx.com/v1/settings/common/symbols", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed struct {
		Status string      `json:"status"`
		Data   []rawSymbol `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if parsed.Status != "ok" {
		return nil, fmt.Errorf("invalid response from HTX")
	}

	log.Printf("🔍 Total received: %d items", len(parsed.Data))
	
	
	var result []model.Symbol
	for _, sym := range parsed.Data {
		if sym.State == "online" && sym.Symbol != "" && sym.TradeEnabled {
			result = append(result, model.Symbol{
				Symbol:        sym.Symbol,
				BaseCurrency:  sym.BaseCurrency,
				QuoteCurrency: sym.QuoteCurrency,
				State:         sym.State,
				TradeEnabled:  sym.TradeEnabled,
			})
		}
	}
	
	return result, nil
}
