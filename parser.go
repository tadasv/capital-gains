package main

import (
	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"
	"log"
	"os"
	"strings"
	"time"
)

type Lot struct {
	DateTime      time.Time
	Symbol        string
	Type          string
	Amount        decimal.Decimal
	Price         decimal.Decimal
	TradeID       string
	TradingFeeUSD decimal.Decimal
}

var (
	decimalSign decimal.Decimal = decimalMustFromString("-1.0")
)

type USD string
type BTC string
type ETH string

func (btc BTC) Decimal() decimal.Decimal {
	if len(btc) == 0 {
		return decimal.Zero
	}

	negative := false
	if strings.Contains(string(btc), "(") {
		negative = true
	}

	v := strings.Replace(string(btc), "(", "", -1)
	v = strings.Replace(v, ")", "", -1)
	v = strings.Replace(v, "BTC", "", -1)
	v = strings.Replace(v, ",", "", -1)
	v = strings.TrimSpace(v)

	d := decimalMustFromString(v)
	if negative {
		d = d.Mul(decimalMustFromString("-1"))
	}
	return d
}

func (eth ETH) Decimal() decimal.Decimal {
	if len(eth) == 0 {
		return decimal.Zero
	}

	negative := false
	if strings.Contains(string(eth), "(") {
		negative = true
	}

	v := strings.Replace(string(eth), "(", "", -1)
	v = strings.Replace(v, ")", "", -1)
	v = strings.Replace(v, "ETH", "", -1)
	v = strings.Replace(v, ",", "", -1)
	v = strings.TrimSpace(v)

	d := decimalMustFromString(v)
	if negative {
		d = d.Mul(decimalMustFromString("-1"))
	}
	return d
}

func (usd USD) Decimal() decimal.Decimal {
	if len(usd) == 0 {
		return decimal.Zero
	}

	negative := false
	if strings.Contains(string(usd), "(") {
		negative = true
	}

	v := strings.Replace(string(usd), "(", "", -1)
	v = strings.Replace(v, ")", "", -1)
	v = strings.Replace(v, "$", "", -1)
	v = strings.Replace(v, ",", "", -1)
	v = strings.TrimSpace(v)

	d := decimalMustFromString(v)
	if negative {
		d = d.Mul(decimalMustFromString("-1"))
	}
	return d
}

func ParseTrades(filename string) ([]*Lot, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	records := []*Record{}
	lots := []*Lot{}

	if err := gocsv.UnmarshalFile(fd, &records); err != nil {
		return nil, err
	}

	for _, r := range records {
		lots = append(lots, r.ToLot())
	}

	return lots, nil
}

// Record is a CSV row from Gemini transaction history.
type Record struct {
	Date                  string `csv:"Date"`
	Time                  string `csv:"Time (UTC)"`
	Type                  string `csv:"Type"`
	Symbol                string `csv:"Symbol"`
	Specification         string `csv:"Specification"`
	LiquidityIndicator    string `csv:"Liquidity Indicator"`
	TradingFeeRate        string `csv:"Trading Fee Rate (bps)"`
	USDAmount             USD    `csv:"USD Amount"`
	TradingFeeUSD         USD    `csv:"Trading Fee (USD)"`
	USDBalance            USD    `csv:"USD Balance"`
	BTCAmount             BTC    `csv:"BTC Amount"`
	TradingFeeBTC         BTC    `csv:"Trading Fee (BTC)"`
	BTCBalance            BTC    `csv:"BTC Balance"`
	ETHAmount             ETH    `csv:"ETH Amount"`
	ETHBalance            ETH    `csv:"ETH Balance"`
	TradeID               string `csv:"Trade ID"`
	OrderID               string `csv:"Order ID"`
	OrderDate             string `csv:"Order Date"`
	OrderTime             string `csv:"Order Time"`
	ClientOrderID         string `csv:"Client Order ID"`
	APISession            string `csv:"API Session"`
	TXHash                string `csv:"TX Hash"`
	DepositTXOutput       string `csv:"Deposit Tx Output"`
	WithdrawalDestination string `csv:"Withdrawal Destination"`
	WithdrawalTXOutput    string `csv:"Withdrawal Tx Output"`
}

func (r *Record) ToLot() *Lot {
	lot := &Lot{
		DateTime:      mustDateParse(r.Date + " " + r.Time),
		Symbol:        r.Symbol,
		Type:          r.Type,
		TradeID:       r.TradeID,
		TradingFeeUSD: r.TradingFeeUSD.Decimal(),
	}

	if r.Type == "Credit" {
		lot.TradeID = "deposit"
		if r.Symbol == symbolUSD {
			// USD deposit
			lot.Amount = r.USDAmount.Decimal()
			lot.Price = decimalMustFromString("1.0")
		} else {
			// Crypto deposit
			if r.Symbol == symbolBTC {
				lot.Amount = r.BTCAmount.Decimal()
			} else if r.Symbol == symbolETH {
				lot.Amount = r.ETHAmount.Decimal()
			} else {
				log.Panicf("credit in unsupported currency: %v", r)
			}
		}
	}

	if r.Type == "Buy" || r.Type == "Sell" {
		if r.Symbol == tradingPairETHUSD {
			lot.Price, lot.Amount = r.calculateETHUSDPriceAndAmount()
		} else if r.Symbol == tradingPairBTCUSD {
			lot.Price, lot.Amount = r.calculateBTCUSDPriceAndAmount()
		} else {
			log.Panicf("unsupported trading pair: %v", r)
		}
	}

	if r.Type == "Debit" {
	}

	return lot
}

func (r *Record) calculateETHUSDPriceAndAmount() (decimal.Decimal, decimal.Decimal) {
	if r.Type == "Buy" {
		// Cost basis
		return r.USDAmount.Decimal().Div(r.ETHAmount.Decimal()).Mul(decimalSign), r.ETHAmount.Decimal()
	} else if r.Type == "Sell" {
		// Sale price
		return r.USDAmount.Decimal().Div(r.ETHAmount.Decimal()).Mul(decimalSign), r.ETHAmount.Decimal().Mul(decimalSign)
	} else {
		// TODO handle credit, pull FMV from a good source
		return decimal.Zero, decimal.Zero
	}
}

func (r *Record) calculateBTCUSDPriceAndAmount() (decimal.Decimal, decimal.Decimal) {
	if r.Type == "Buy" {
		// Cost basis
		return r.USDAmount.Decimal().Div(r.BTCAmount.Decimal()).Mul(decimalSign), r.BTCAmount.Decimal()
	} else if r.Type == "Sell" {
		// Sale price
		return r.USDAmount.Decimal().Div(r.BTCAmount.Decimal()).Mul(decimalSign), r.BTCAmount.Decimal().Mul(decimalSign)
	} else {
		// TODO handle credit, pull FMV from a good source
		return decimal.Zero, decimal.Zero
	}
}

func decimalMustFromString(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustDateParse(v string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", v)
	if err != nil {
		panic(err)
	}
	return t
}
