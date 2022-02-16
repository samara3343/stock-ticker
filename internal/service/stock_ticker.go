package service

import (
	"context"
	"encoding/json"
	"log"
	"stock-ticker/internal/dto"
	"sync"
	"time"

	"github.com/SimpleApplicationsOrg/stock/alphavantage"
	"github.com/coocood/freecache"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type (
	stockTimeSeriesGetter interface {
		TimeSeries(string, string) (*alphavantage.TimeSeriesData, error)
	}
	StockTickerConfig struct {
		StockSymbol string
		NDays       int
	}
)

type StockTickerService struct {
	cfg                   StockTickerConfig
	stockTimeSeriesGetter stockTimeSeriesGetter
	stockTimeSeriesCache  *freecache.Cache
	stockTimeSeriesMu     *sync.Mutex
}

func NewStockTickerService(cfg StockTickerConfig, stockTimeSeriesGetter stockTimeSeriesGetter) *StockTickerService {
	return &StockTickerService{
		cfg:                   cfg,
		stockTimeSeriesGetter: stockTimeSeriesGetter,
		stockTimeSeriesCache:  freecache.NewCache(64 * 1024 * 1024 * 1024),
		stockTimeSeriesMu:     &sync.Mutex{},
	}
}

func (svc *StockTickerService) getStockTimeSeriesData() (*alphavantage.TimeSeriesData, error) {
	cacheKey := "TIME_SERIES_DAILY-" + svc.cfg.StockSymbol

	getFromCache := func() (*alphavantage.TimeSeriesData, error) {
		cachedTimeSeriesBytes, err := svc.stockTimeSeriesCache.Get([]byte(cacheKey))
		if err != nil {
			return nil, err
		}
		var result alphavantage.TimeSeriesData
		if err := json.Unmarshal(cachedTimeSeriesBytes, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	// try cache first
	result, err := getFromCache()
	if err == nil {
		return result, nil
	}

	// we don't want to bombard the API
	// all the routines should wait until the cache is populated and get data from the cache
	svc.stockTimeSeriesMu.Lock()
	defer svc.stockTimeSeriesMu.Unlock()

	// it's possible that cache has already been updated, try ir one more time
	result, err = getFromCache()
	if err == nil {
		return result, nil
	}

	tss, err := svc.stockTimeSeriesGetter.TimeSeries("TIME_SERIES_DAILY", svc.cfg.StockSymbol)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get time series data")
	}

	cachedTimeSeriesBytes, err := json.Marshal(tss)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal data returned from Alpha Vantage API")
	}
	tz, err := time.LoadLocation(tss.MetaData.TimeZone())
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse timezone")
	}
	eodTmr := time.Now().In(tz).Add(24 * time.Hour).Truncate(24 * time.Hour).Add(-1)
	log.Printf("saving cache, key=%s, expiresAt=%s, tz=%s", cacheKey, eodTmr, tz.String())
	if err := svc.stockTimeSeriesCache.Set([]byte(cacheKey), cachedTimeSeriesBytes, int(time.Until(eodTmr).Seconds())); err != nil {
		return nil, errors.Wrap(err, "unable to store data in cache")
	}

	return tss, nil
}

func (svc *StockTickerService) ListClosingStockPricesInfo(ctx context.Context) (*dto.GetClosingStockPriceInfo, error) {
	_ = ctx

	tss, err := svc.getStockTimeSeriesData()
	if err != nil {
		return nil, err
	}
	if tss.TimeSeries == nil {
		return nil, errors.Errorf("unable to get time series for %s", svc.cfg.StockSymbol)
	}
	timeSeries := *tss.TimeSeries

	truncateStrSlice := func(ss []string, n int) []string {
		if len(ss) >= n {
			return ss[:n]
		}
		return ss
	}

	dates := truncateStrSlice(tss.TimeSeries.TimeStamps(), svc.cfg.NDays)
	result := dto.GetClosingStockPriceInfo{}
	for _, ts := range dates {
		closingPrice, err := decimal.NewFromString(timeSeries[ts].Close())
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse closing price")
		}
		result.Prices = append(result.Prices, dto.ClosingStockPrice{
			Date:  ts,
			Price: closingPrice,
		})
		result.AveragePrice = result.AveragePrice.Add(closingPrice)
	}
	result.AveragePrice = result.AveragePrice.DivRound(decimal.NewFromInt(int64(len(dates))), 4)

	return &result, nil
}
