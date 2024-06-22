package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/donohutcheon/valr-go"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Function to check if file exists and return a bool.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func maybeLoadEnvFile() {
	if !fileExists(".env") {
		return
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %s", err.Error())
	}
}

func main() {
	maybeLoadEnvFile()
	// Create a channel to receive OS signals.
	sigs := make(chan os.Signal, 1)

	// Notify the channel of SIGINT (Ctrl+C) and SIGTERM signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to notify the main goroutine when it's time to exit.
	done := make(chan bool, 1)

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to wait for the signal.
	go func() {
		<-sigs
		done <- true
		cancel()
	}()

	go pollMarketsForever(ctx)
	go listSupportedPairs(ctx)

	// Block the main goroutine until a signal is received.
	<-done
	fmt.Println("Exiting...")
}

func pollMarketsForever(ctx context.Context) {
	client := valr.NewClient()
	client.SetAuth(os.Getenv("VA_KEY_ID"), os.Getenv("VA_SECRET"))
	endTime := time.Now()
	startTime := endTime.Add(-6 * time.Minute)
	req := &valr.GetAuthTradeHistoryForPairRequest{
		Pair:      "BTCZAR",
		Limit:     100,
		Skip:      0,
		StartTime: startTime,
		EndTime:   endTime,
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := client.GetAuthTradeHistoryForPairRequest(ctx, req)
			if errors.Is(err, valr.ErrTooManyRequests) {
				log.Fatal(err)
			}
			if err != nil {
				fmt.Println(err)
				return
			}

			if len(resp) == 0 {
				continue
			}
			fmt.Printf("got trades, %d, %d, %d\n", resp[0].SequenceID, resp[len(resp)-1].SequenceID, len(resp))
			if len(resp) < 100 {
			} else {
				req.Skip += 100
			}

			for _, trade := range resp {
				fmt.Printf("Trade:\n\tPair: %s\n\tTaker's Side: %s\n\tPrice: %s\n\tQuantity: %s\n\tTimestamp: %s\n\tSequence: %d\n\tTrade ID: %s\n", trade.Pair, trade.TakerSide, trade.Price, trade.Quantity, trade.TradedAt, trade.SequenceID, trade.ID)
			}
		}
		return
	}
}

func listSupportedPairs(ctx context.Context) {
	client := valr.NewClient()
	req := &valr.GetCurrencyPairsByTypeRequest{
		PairTpe: valr.PairTypeSpot,
	}

	resp, err := client.GetCurrencyPairsByType(ctx, req)
	if errors.Is(err, valr.ErrTooManyRequests) {
		log.Fatal(err)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, pi := range resp {
		i, err := strconv.ParseInt(pi.BaseDecimalPlaces, 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(
			"Pair Info:\n\tbase: %s\n\tcounter: %s\n\tshort name: %s\n\t"+
				"pair type: %s\n\tsymbol: %s\n\tactive: %#v\n\tbase decimal places: %d\n",
			pi.BaseCurrency, pi.QuoteCurrency, pi.ShortName, pi.CurrencyPairType, pi.Symbol,
			pi.Active, i,
		)
	}

}
