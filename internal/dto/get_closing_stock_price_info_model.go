package dto

import (
	"github.com/shopspring/decimal"
)

type ClosingStockPrice struct {
	Date  string          `json:"date"`
	Price decimal.Decimal `json:"price"`
}

type GetClosingStockPriceInfo struct {
	Prices       []ClosingStockPrice `json:"prices"`
	AveragePrice decimal.Decimal     `json:"averagePrice"`
}
