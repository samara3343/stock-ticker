package handler

import (
	"net/http"
	"stock-ticker/internal/service"

	"github.com/gin-gonic/gin"
)

func stockClosingPriceInfoHandler(stockTickerService *service.StockTickerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cspi, err := stockTickerService.ListClosingStockPricesInfo(c.Request.Context())
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, cspi)
	}
}

func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, "OK")
	}
}
