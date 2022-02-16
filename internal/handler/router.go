package handler

import (
	"stock-ticker/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(stockTickerService *service.StockTickerService) *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/stock-closing-price-info", stockClosingPriceInfoHandler(stockTickerService))
	router.GET("/api/health", healthHandler())

	return router
}
