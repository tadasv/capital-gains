# Bitcoin capital gains calculator

Disclaimer: I'm not a CPA or otherwise authorized to give tax advice. This code
is released under the GPL, which in particular disavows all responsibility for
the accuracy of these results.

## About

This tool lets you calculate capital gains on trades made at Gemini
(http://gemini.com). To calculate capital gains, FIFO method is used. This
means that any sale can span one or more buy lots. Crypto deposits in BTC and
ETH are treated as buys at FMV at the date of the deposit was received buy the
exchange.

For every buy trade, cost basis is derived from historical trade data provided
in the Gemini export file.

Transaction fees are not taken into account when calculating capital gains.

Fair market price is retrieved from the following sources:

- BTC: https://api.blockchain.info/charts/market-price?timespan=2years&format=csv
- ETH: https://etherscan.io/chart/etherprice?output=csv

## Use

Download transaction history from Gemini. You should receive file in `.xlsx`.
Convert this file to a CSV and remove last row containing transaction summary.

```
$ capital-gains -filename=<csv file>

USD:
	capital gains: 0
ETHUSD:
	capital gains: 1338.055061499999999976266685
BTCUSD:
	capital gains: 0
ETH:
	capital gains: 0
```

Running the command above will produce a CSV file in the workdir for each trading pair with calculations, e.g.:

```
DateTime,TradeID,Type,Symbol,Amount,CostBasis,Price,CapitalGain
2017-06-14T23:59:03.017Z,83,Buy,ETH,1.45791615,0,342.0978634470850741,0
2017-06-20T03:16:45.755Z,42,Sell,ETH,0,362.0030234938089712,342.0978634470850741,29.0200513146795171644436
2017-06-21T14:35:31.385Z,9,Buy,ETH,1,0,320,0
2017-08-30T23:41:03.521Z,6,Buy,ETH,1.2945,0,385.3225183468520664,0
2017-12-07T11:15:46.581Z,168,Buy,ETH,5,0,410,0
2017-12-08T04:35:36.956Z,974,Buy,ETH,5,0,422.5,0
2017-12-10T23:14:16.66Z,2429,Buy,ETH,7,0,430,0
2017-12-11T01:20:01.825Z,deposit,Credit,ETH,0.17292208,0,513.29,0
2017-12-13T14:06:00.877Z,964,Buy,ETH,0.726346,0,687.0004102727901028,0
2017-12-13T18:20:33.993Z,963,Buy,ETH,0.548529,0,680.0005104561472593,0
2017-12-18T15:46:08.42Z,deposit,Credit,ETH,0.05192183,0,785.99,0
2017-12-24T14:39:34.642Z,972,Sell,ETH,0.00000015,342.0978634470850741,653.9991771985189573,0.00004678519706271508248
2017-12-24T14:39:34.642Z,972,Sell,ETH,1,320,653.9991771985189573,333.9991771985189573
2017-12-24T14:39:34.642Z,72,Sell,ETH,1.2945,385.3225183468520664,653.9991771985189573,347.80193488348279027005
2017-12-24T14:39:34.642Z,72,Sell,ETH,0,653.9991771985189573,410,660.137541318121672526690705
2017-12-24T14:39:34.672Z,80,Sell,ETH,0,1111.1111111111111111,410,0.0063099999999999999999
```
