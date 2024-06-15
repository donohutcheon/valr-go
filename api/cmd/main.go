package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/donohutcheon/valr-go/api"
	"github.com/joho/godotenv"
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
	ctx := context.Background()
	maybeLoadEnvFile()

	client := api.NewClient()
	client.SetAuth(os.Getenv("VA_KEY_ID"), os.Getenv("VA_SECRET"))
	startTime := time.Date(2022, 5, 22, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2022, 5, 23, 0, 0, 0, 0, time.UTC)

	fmt.Println("Default string ", startTime.String())

	req := &api.GetAuthTradeHistoryForPairRequest{
		Pair:      "BTCZAR",
		Limit:     100,
		Skip:      0,
		StartTime: startTime,
		EndTime:   endTime,
	}
	for {
		resp, err := client.GetAuthTradeHistoryForPairRequest(ctx, req)
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(*resp) < 100 {
			break
		} else {
			req.Skip += 100
		}

		for _, trade := range *resp {
			fmt.Printf("%+v\n", trade)
		}
		time.Sleep(time.Millisecond * 200)
	}

	fmt.Println("done")
}
