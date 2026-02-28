# Triangular Arbitrage Scanner

Polls Coinbase spot prices concurrently for multiple trading pairs,
builds an in-memory rate graph, and runs DFS path enumeration to
detect cross-currency arbitrage opportunities.

## How It Works

Every 5 seconds, the scanner fetches spot prices for BTC and
selected altcoins against USD and BTC. Rates are stored in a
directed graph. A recursive DFS walks all paths up to a
configurable depth, looking for cycles that return more than
the starting amount after fees.

## Run

go run .

## Configuration

Altcoins, polling interval, fee rate, and profit threshold
are currently set in main.go.