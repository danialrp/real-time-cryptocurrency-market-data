package utils

import (
	"fmt"
	"strconv"
)

func CalculateAverageFromStr(strOne, strTwo string) (float64, error) {
	fltOne, err := strconv.ParseFloat(strOne, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid buy rate: %v", err)
	}

	fltTwo, err := strconv.ParseFloat(strTwo, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid sell rate: %v", err)
	}

	average := (fltOne + fltTwo) / 2.0
	return average, nil
}