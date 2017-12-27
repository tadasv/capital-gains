package main

import (
	"encoding/csv"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
	"time"
)

const marketPriceDataURL = "https://api.blockchain.info/charts/market-price?timespan=2years&format=csv"
const ethMarketPriceDataURL = "https://etherscan.io/chart/etherprice?output=csv"

type FMV struct {
	Date  string
	Price decimal.Decimal
}

type MarketData struct {
	Prices map[string]*FMV
}

func GetBTCMarketData() (*MarketData, error) {
	resp, err := http.Get(marketPriceDataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	marketData := &MarketData{
		Prices: make(map[string]*FMV),
	}

	reader := csv.NewReader(resp.Body)
	record, err := reader.Read()

	for err == nil {
		date, err := time.Parse("2006-01-02 00:00:00", record[0])
		if err != nil {
			panic(err)
		}

		fmv := &FMV{
			Date:  date.Format("2006-01-02"),
			Price: decimalMustFromString(record[1]),
		}

		marketData.Prices[fmv.Date] = fmv

		record, err = reader.Read()
		if len(record) == 0 {
			break
		}
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return marketData, nil
}

func GetETHMarketData() (*MarketData, error) {
	resp, err := http.Get(ethMarketPriceDataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	marketData := &MarketData{
		Prices: make(map[string]*FMV),
	}

	reader := csv.NewReader(resp.Body)
	record, err := reader.Read()
	// to skip the header
	first := true

	for err == nil {
		if !first {
			date, err := time.Parse("1/2/2006", record[0])
			if err != nil {
				panic(err)
			}

			fmv := &FMV{
				Date:  date.Format("2006-01-02"),
				Price: decimalMustFromString(record[2]),
			}

			marketData.Prices[fmv.Date] = fmv
		}

		first = false
		record, err = reader.Read()
		if len(record) == 0 {
			break
		}
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return marketData, nil
}
