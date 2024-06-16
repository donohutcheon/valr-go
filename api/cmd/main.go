package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/donohutcheon/valr-go/api"
	"github.com/donohutcheon/valr-go/api/streaming"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
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
		sig := <-sigs
		fmt.Println()
		fmt.Println("Received signal:", sig)
		done <- true
		cancel()
	}()

	go streamMarketsForever(ctx)
	go pollMarketsForever(ctx)

	// Block the main goroutine until a signal is received.
	fmt.Println("Awaiting signal")
	<-done
	fmt.Println("Exiting")
}

func pollMarketsForever(ctx context.Context) {
	client := api.NewClient()
	client.SetAuth(os.Getenv("VA_KEY_ID"), os.Getenv("VA_SECRET"))
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	req := &api.GetAuthTradeHistoryForPairRequest{
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
			if errors.Is(err, api.ErrTooManyRequests) {
				log.Fatal(err)
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(*resp) < 100 {
				fmt.Println("done")
				return
			} else {
				req.Skip += 100
			}

			for _, trade := range *resp {
				fmt.Printf("trade %+v\n", trade)
			}
		}
	}
}

func streamMarketsForever(ctx context.Context) {
	c, err := streaming.Dial(
		os.Getenv("VA_KEY_ID"),
		os.Getenv("VA_SECRET"),
		streaming.WithUpdateCallback(func(update streaming.MessageTradeUpdate) {
			fmt.Printf("Trade Update Callback: %+v\n", update)
		}),
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
