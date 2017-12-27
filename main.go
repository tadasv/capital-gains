package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	symbolUSD         = "USD"
	symbolETH         = "ETH"
	symbolBTC         = "BTC"
	tradingPairETHUSD = "ETHUSD"
	tradingPairBTCUSD = "BTCUSD"
)

func main() {
	fileName := flag.String("filename", "history.csv", "trade history csv")
	flag.Parse()

	btcMarketData, err := GetBTCMarketData()
	if err != nil {
		panic(err)
	}

	ethMarketData, err := GetETHMarketData()
	if err != nil {
		panic(err)
	}

	cr := NewCompositeReport(btcMarketData, ethMarketData)

	lots, err := ParseTrades(*fileName)
	if err != nil {
		panic(err)
	}

	for _, lot := range lots {
		cr.Add(lot)
	}

	fmt.Printf("%s\n", cr)

	for key, report := range cr.Reports {
		if len(report.Calculations) > 0 {
			fd, err := os.OpenFile(key+".csv", os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				panic(err)
			}

			report.WriteCalculations(fd)
			fd.Close()
		}
	}
}
