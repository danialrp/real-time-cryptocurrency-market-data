package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"irbtc_streamer/internal/redis"
	"irbtc_streamer/internal/utils"
)

const (
	usdtPriceIrbtcURL = "https://base.irbtc.com/api/v1/coins/symbol/usdt"
	usdtPriceRaterURL = "https://rater.irbtc.net/api/prices/latest?symbol=usdt"
)

type usdtIrbtcResponse struct {
	Data struct {
		BuyRate  string `json:"buy_rate"`
		SellRate string `json:"sell_rate"`
	} `json:"data"`
}

type usdtRaterResponse struct {
	Data []struct {
		BuyRate  string `json:"buy_price"`
		SellRate string `json:"sell_price"`
		LastRate string `json:"last_price"`
	} `json:"data"`
}

func FetchAndStoreUSDTPriceFromIrbtc() error {
	resp, err := http.Get(usdtPriceIrbtcURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	var result usdtIrbtcResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	avg, err := utils.CalculateAverageFromStr(result.Data.SellRate, result.Data.BuyRate)
	if err != nil {
		log.Println("error calculating average:", err)
		avg, _ = strconv.ParseFloat(result.Data.BuyRate, 64)
	}

	payload := map[string]string{
		"buy_rate":   result.Data.BuyRate,
		"sell_rate":  result.Data.SellRate,
		"last_rate":  fmt.Sprintf("%.0f", avg),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	key := "irbtc:price.irbtc:usdtirt"
	data, _ := json.Marshal(payload)

	return redis.ClientGeneral.Set(context.Background(), key, data, 600*time.Second).Err()
}

func FetchAndStoreUSDTPriceFromRater() error {
	resp, err := http.Get(usdtPriceRaterURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	var result usdtRaterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	payload := map[string]string{
		"buy_rate":   result.Data[0].BuyRate,
		"sell_rate":  result.Data[0].SellRate,
		"last_rate":  result.Data[0].LastRate,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	key := "irbtc:price.rater:usdtirt"
	data, _ := json.Marshal(payload)

	return redis.ClientGeneral.Set(context.Background(), key, data, 600*time.Second).Err()
}
