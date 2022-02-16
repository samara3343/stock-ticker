package cmd

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"stock-ticker/internal/handler"
	"stock-ticker/internal/service"
	"strconv"
	"syscall"
	"time"

	"github.com/SimpleApplicationsOrg/stock/alphavantage"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func startServer() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	gin.SetMode(viper.GetString("gin-mode"))

	avclient, err := alphavantage.NewAVClient()
	if err != nil {
		log.Fatalf("FATAL: unable to initialize alpha vantage client, error=%v", err)
	}
	stockTickerService := service.NewStockTickerService(
		service.StockTickerConfig{
			StockSymbol: viper.GetString("query-stock-symbol"),
			NDays:       viper.GetInt("query-ndays"),
		}, avclient)

	router := handler.NewRouter(stockTickerService)

	srv := &http.Server{
		Addr:              net.JoinHostPort(viper.GetString("host"), strconv.Itoa(viper.GetInt("port"))),
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    128 * 1024 * 1024,
	}

	log.Printf("INFO: server starting at %s", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("unable to start server, error=%v", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("INFO: shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("FATAL: server forced to shutdown: ", err)
	}

	log.Println("INFO: server exiting")
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts an HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("host", "0.0.0.0", "a server host")
	serveCmd.PersistentFlags().Int32("port", 8080, "a server port")
	serveCmd.PersistentFlags().String("gin-mode", gin.ReleaseMode, "GIN mode, release or debug. Alternatively env variable GIN_MODE can be used.")
	serveCmd.PersistentFlags().String("query-stock-symbol", "MSFT", "A stock symbol that is used for querying the price information.")
	serveCmd.PersistentFlags().Int("query-ndays", 7, "A number of days to query for from the stock API.")
	serveCmd.PersistentFlags().String("alpha-vantage-url", "https://www.alphavantage.co", "A URL for Alpha Vantage API.")
	serveCmd.PersistentFlags().String("alpha-vantage-key-name", "apikey", "An API for Alpha Vantage API.")
	serveCmd.PersistentFlags().String("alpha-vantage-key-value", "", "An API key for Alpha Vantage API.")

	_ = viper.BindPFlag("host", serveCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	_ = viper.BindPFlag("gin-mode", serveCmd.PersistentFlags().Lookup("gin-mode"))
	_ = viper.BindPFlag("query-stock-symbol", serveCmd.PersistentFlags().Lookup("query-stock-symbol"))
	_ = viper.BindPFlag("query-ndays", serveCmd.PersistentFlags().Lookup("query-ndays"))

	_ = viper.BindEnv("gin-mode", "GIN_MODE")
	_ = viper.BindEnv("query-stock-symbol", "QUERY_STOCK_SYMBOL")
	_ = viper.BindEnv("query-ndays", "QUERY_NDAYS")
	_ = viper.BindEnv("alpha-vantage-url", "ALPHA_VANTAGE_URL")
	_ = viper.BindEnv("alpha-vantage-key-name", "ALPHA_VANTAGE_KEY_NAME")
	_ = viper.BindEnv("alpha-vantage-key-value", "ALPHA_VANTAGE_KEY_VALUE")
}
