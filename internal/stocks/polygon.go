package stocks

// package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

func main() {
	c := polygon.New(os.Getenv("POLYGON_KEY"))

	currentTime := time.Now()
	yesterday := time.Now().AddDate(0, 0, -1)

	estTZ, err := time.LoadLocation("America/New_York")

	if err != nil {
		panic(err)
	}

	yesterdayClose := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		16,
		0,
		0,
		0,
		estTZ,
	)

	if err != nil {
		panic(err)
	}

	params := models.ListAggsParams{
		Ticker:     "SPY",
		Multiplier: 1,
		Timespan:   "day",
		From:       models.Millis(yesterdayClose),
		To:         models.Millis(currentTime),
	}.WithOrder(models.Order("asc")).WithLimit(120)

	iter := c.ListAggs(context.Background(), params)

	for iter.Next() {
		fmt.Println(iter.Item())
	}

	if iter.Err() != nil {
		log.Fatal(iter.Err())
	}

	fmt.Println("PREVIOUS DAY")
	fmt.Println("============")

	prevParams := models.GetPreviousCloseAggParams{
		Ticker: "SPY",
	}.WithAdjusted(true)

	res, err := c.GetPreviousCloseAgg(context.Background(), prevParams)

	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
