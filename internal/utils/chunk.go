package utils

import (
	"irbtc_streamer/internal/model"
)

func ChunkSymbols(symbols []model.Symbol, chunkSize int) [][]model.Symbol {
	var chunks [][]model.Symbol
	for chunkSize < len(symbols) {
		symbols, chunks = symbols[chunkSize:], append(chunks, symbols[0:chunkSize:chunkSize])
	}
	return append(chunks, symbols)
}
