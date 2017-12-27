package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateCapitalGainsWithoutDeposits(t *testing.T) {
	tests := []struct {
		lots         []*Lot
		expectedGain string
	}{
		{
			lots: []*Lot{
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.2"),
					Price:  decimalMustFromString("300"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.2"),
					Price:  decimalMustFromString("400"),
				},
			},
			expectedGain: "20",
		},
		{
			lots: []*Lot{
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("100"),
				},
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("200"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.2"),
					Price:  decimalMustFromString("220"),
				},
			},
			expectedGain: "14",
		},
		{
			lots: []*Lot{
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.3"),
					Price:  decimalMustFromString("100"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("200"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("150"),
				},
			},
			expectedGain: "15",
		},
		{
			lots: []*Lot{
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("100"),
				},
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("50"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.05"),
					Price:  decimalMustFromString("150"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.15"),
					Price:  decimalMustFromString("150"),
				},
			},
			expectedGain: "15",
		},
		{
			lots: []*Lot{
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("100"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.05"),
					Price:  decimalMustFromString("150"),
				},
				&Lot{
					Type:   "Buy",
					Amount: decimalMustFromString("0.1"),
					Price:  decimalMustFromString("50"),
				},
				&Lot{
					Type:   "Sell",
					Amount: decimalMustFromString("0.15"),
					Price:  decimalMustFromString("150"),
				},
			},
			expectedGain: "15",
		},
	}

	for _, test := range tests {
		r := NewRunningReport(nil, nil)
		for _, lot := range test.lots {
			r.Add(lot)
		}
		assert.True(t, r.TotalCapitalGain.Equal(decimalMustFromString(test.expectedGain)))
	}
}

func TestCalculateCapitalGainsWithTradingFees(t *testing.T) {
	lots := []*Lot{
		&Lot{
			Type:          "Buy",
			Amount:        decimalMustFromString("0.2"),
			Price:         decimalMustFromString("300"),
			TradingFeeUSD: decimalMustFromString("-1"),
		},
		&Lot{
			Type:          "Sell",
			Amount:        decimalMustFromString("0.2"),
			Price:         decimalMustFromString("400"),
			TradingFeeUSD: decimalMustFromString("-1"),
		},
	}
	expectedGain := "18"

	r := NewRunningReport(nil, nil)
	for _, lot := range lots {
		r.Add(lot)
	}
	assert.True(t, r.TotalCapitalGain.Equal(decimalMustFromString(expectedGain)))
}
