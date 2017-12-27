package main

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"time"
)

func NewCompositeReport(btcMarketData *MarketData, ethMarketData *MarketData) *CompositeRunningReport {
	return &CompositeRunningReport{
		Reports:       make(map[string]*RunningReport),
		BTCMarketData: btcMarketData,
		ETHMarketData: ethMarketData,
	}
}

func (cr *CompositeRunningReport) String() string {
	s := ""

	for key, value := range cr.Reports {
		s += fmt.Sprintf("%s:\n\t%s\n", key, value)
	}

	return s
}

func (cr *CompositeRunningReport) Add(lot *Lot) {
	symbol := lot.Symbol
	if lot.Type == "Credit" {
		// Make sure to assign FMV price for cryto currency deposits if value is not already set.
		if lot.Price.Equal(decimal.Zero) {
			// Get FMV
			var md *MarketData
			if lot.Symbol == symbolBTC {
				md = cr.BTCMarketData
			} else if lot.Symbol == symbolETH {
				md = cr.ETHMarketData
			} else {
				log.Fatalf("unsupported symbol: %v", lot)
			}

			fmvPrice, ok := md.Prices[lot.DateTime.Format("2006-01-02")]
			if !ok {
				log.Fatalf("failed to get fmv: %v", lot)
			} else {
				lot.Price = fmvPrice.Price
			}
		}

		// Treat crypto deposits as trading pairs at FMV of the deposit date.
		if symbol == symbolBTC {
			symbol = tradingPairBTCUSD
		}

		if symbol == symbolETH {
			symbol = tradingPairETHUSD
		}
	}

	r, ok := cr.Reports[symbol]
	if !ok {
		r = NewRunningReport(cr.BTCMarketData, cr.ETHMarketData)
		cr.Reports[symbol] = r
	}

	r.Add(lot)
}

type CompositeRunningReport struct {
	// trading pair -> report map
	Reports       map[string]*RunningReport
	BTCMarketData *MarketData
	ETHMarketData *MarketData
}

func NewRunningReport(btcMarketData *MarketData, ethMarketData *MarketData) *RunningReport {
	return &RunningReport{
		FIFOBuys:      make([]*Lot, 0),
		BTCMarketData: btcMarketData,
		ETHMarketData: ethMarketData,
		Calculations:  make([]*CalculationRecord, 0),
	}
}

type RunningReport struct {
	BTCMarketData    *MarketData
	ETHMarketData    *MarketData
	TotalCapitalGain decimal.Decimal
	FIFOBuys         []*Lot
	Calculations     []*CalculationRecord
}

func (r *RunningReport) WriteCalculations(w io.Writer) error {
	return gocsv.Marshal(r.Calculations, w)
}

func (r *RunningReport) Add(lot *Lot) {
	if lot.Type == "Buy" || lot.Type == "Credit" {
		r.addBuy(lot)
		return
	} else if lot.Type == "Sell" {
		r.addSell(lot)
		return
	}

	log.Printf("skipping unsupported lot type: %v", lot)
}

func (r *RunningReport) addBuy(lot *Lot) {
	r.FIFOBuys = append(r.FIFOBuys, lot)
	r.TotalCapitalGain = r.TotalCapitalGain.Add(lot.TradingFeeUSD)
	r.Calculations = append(r.Calculations, &CalculationRecord{
		DateTime: lot.DateTime,
		TradeID:  lot.TradeID,
		Type:     lot.Type,
		Symbol:   lot.Symbol[:3],
		Amount:   lot.Amount,
		Price:    lot.Price,
	})
}

func (r *RunningReport) addSell(lot *Lot) {
	for lot.Amount.GreaterThan(decimal.Zero) {
		firstBuy := r.FIFOBuys[0]
		if firstBuy.Amount.GreaterThanOrEqual(lot.Amount) {
			// Here firstBuy covers entire sale. So we just calculate capital gain.
			capitalGain := r.calculateGain(firstBuy, lot)
			firstBuy.Amount = firstBuy.Amount.Sub(lot.Amount)
			lot.Amount = decimal.Zero

			r.Calculations = append(r.Calculations, &CalculationRecord{
				DateTime:    lot.DateTime,
				TradeID:     lot.TradeID,
				Type:        lot.Type,
				Symbol:      lot.Symbol[:3],
				Amount:      lot.Amount,
				CostBasis:   lot.Price,
				Price:       firstBuy.Price,
				CapitalGain: capitalGain,
			})

			r.TotalCapitalGain = r.TotalCapitalGain.Add(capitalGain)
		} else {
			// Here firstBuy does not cover entire sale. Fill as much as we
			// can and move to the next buy lot in the queue.

			// clone lot and adjust amount to match buy amount
			sell := *lot
			sell.Amount = firstBuy.Amount
			capitalGain := r.calculateGain(firstBuy, &sell)

			r.Calculations = append(r.Calculations, &CalculationRecord{
				DateTime:    lot.DateTime,
				TradeID:     lot.TradeID,
				Type:        lot.Type,
				Symbol:      lot.Symbol[:3],
				Amount:      sell.Amount,
				CostBasis:   firstBuy.Price,
				Price:       sell.Price,
				CapitalGain: capitalGain,
			})

			lot.Amount = lot.Amount.Sub(firstBuy.Amount)
			firstBuy.Amount = decimal.Zero

			r.TotalCapitalGain = r.TotalCapitalGain.Add(capitalGain)
		}

		if firstBuy.Amount.Equal(decimal.Zero) {
			// pop fully processed buy record from the queue
			r.FIFOBuys = r.FIFOBuys[1:]
		}
	}

	// Trading feeds are negative so we just add it to the gain
	r.TotalCapitalGain = r.TotalCapitalGain.Add(lot.TradingFeeUSD)
}

func (r *RunningReport) calculateGain(buy *Lot, sell *Lot) decimal.Decimal {
	// TODO separate long vs short gains
	capitalGain := sell.Price.Sub(buy.Price).Mul(sell.Amount)
	return capitalGain
}

func (r *RunningReport) String() string {
	return fmt.Sprintf("capital gains: %s", r.TotalCapitalGain)
}

type CalculationRecord struct {
	DateTime    time.Time
	TradeID     string
	Type        string
	Symbol      string
	Amount      decimal.Decimal
	CostBasis   decimal.Decimal
	Price       decimal.Decimal
	CapitalGain decimal.Decimal
}
