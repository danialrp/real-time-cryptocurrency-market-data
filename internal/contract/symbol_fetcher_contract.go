package contract

import "time"

type SymbolFetcher interface {
	StartSymbolFetcher(interval time.Duration)
}