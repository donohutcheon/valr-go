package main

import (
	"context"
	"fmt"
	"github.com/donohutcheon/valr-go/api/streaming"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	go streamMarketsForever(ctx)

	// Block the main goroutine until a signal is received.
	<-done
}

func streamMarketsForever(ctx context.Context) {
	c, err := streaming.Dial(
		os.Getenv("VA_KEY_ID"),
		os.Getenv("VA_SECRET"),
		streaming.WithUpdateCallback(tradeUpdateCallback(ctx)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	c.SubscribeToMarkets([]string{"BTCZAR", "ETHZAR", "SOLZAR"})
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

func tradeUpdateCallback(_ context.Context) streaming.UpdateCallback {
	return func(update streaming.MessageTradeUpdate) {
		fmt.Printf("Trade:\n\tPair: %s\n\tTaker's Side: %s\n\tPrice: %s\n\tQuantity: %s\n\tTimestamp: %s\n\tSequence: %s\n\tID: %s\n",
			update.CurrencyPairSymbol,
			update.Data.TakerSide,
			update.Data.Price,
			update.Data.Quantity,
			update.Data.TradedAt,
			"<Not available>",
			update.Data.ID)
	}
}
