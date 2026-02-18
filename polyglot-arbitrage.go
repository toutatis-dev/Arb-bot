package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Exchange interface {
	GetPrice(pair string) (float64, error)
	GetName() string
}

type Graph struct {
	Rates map[string]map[string]float64 // map of a map - first string is base currency, second is converted currency, float is exchange rate

}

type MockExchange struct {
	Pair  string
	Price float64
}

type Coinbase struct {
}

func (c Coinbase) GetName() string {
	return "Coinbase"
}

type CoinbaseResponse struct {
	Data struct {
		Base     string `json:"base"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	} `json:"data"`
}

type Binance struct {
	PairMapping map[string]string
}

func (b Binance) GetName() string {
	return "Binance"
}

type BinanceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func main() {

	coinbase := Coinbase{}

	ticker := time.NewTicker(5 * time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	fmt.Printf("Bot started, Ctrl + C to stop\n")

	market := NewGraph()

	for {
		select {
		case <-ticker.C:
			btcPrice, _ := coinbase.GetPrice("BTC-USD")
			ethPrice, _ := coinbase.GetPrice("ETH-USD")

			market.AddRate("USD", "BTC", 1/btcPrice)
			market.AddRate("BTC", "USD", btcPrice)
			market.AddRate("ETH", "USD", 1/ethPrice)
			market.AddRate("USD", "ETH", ethPrice)

			CalculateDynamicPath(market, 100.0, "USD", 100.0, []string{"USD"}, 3)

		case <-quit:
			fmt.Println("\n Quitting\n")
			ticker.Stop()
			fmt.Println("Goodbye!")
			return
		}
	}
}

func NewGraph() *Graph {

	return &Graph{Rates: make(map[string]map[string]float64)}

}

func (g *Graph) AddRate(source string, destination string, rate float64) {

	if g.Rates[source] == nil {

		g.Rates[source] = make(map[string]float64)
	}

	g.Rates[source][destination] = rate

}

func CalculateDynamicPath(g *Graph, startingamount float64, currentNode string, currentAmount float64, path []string, maxSteps int) {

	if len(path) > maxSteps {
		return //exceeded max steps
	}

	if len(path) > 1 && path[0] == currentNode {
		if currentAmount > startingamount {
			fmt.Printf("Potential Arbitrage found!\n Profit made %.2f\n", currentAmount-startingamount)
			fmt.Printf("Path found %v\n", path)
		}
		return
	}

	for nextNode, rate := range g.Rates[currentNode] {

		newAmount := currentAmount * rate
		newPath := append([]string{}, path...)
		newPath = append(newPath, nextNode)

		CalculateDynamicPath(g, startingamount, nextNode, newAmount, newPath, maxSteps)
	}

}

func (c Coinbase) GetPrice(pair string) (float64, error) {

	url := fmt.Sprintf("https://api.coinbase.com/v2/prices/%s/spot", pair)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("API returned non 200 status")
	}

	var cbResponse CoinbaseResponse

	if err := json.NewDecoder(resp.Body).Decode(&cbResponse); err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(cbResponse.Data.Amount, 64)
	if err != nil {
		return 0, errors.New("Failed to convert price")
	}

	return price, nil
}

func (b Binance) GetPrice(pair string) (float64, error) {

	binanceSymbol, ok := b.PairMapping[pair]
	if !ok {
		return 0, errors.New("The mapping for Binance isnt configured")
	}

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", binanceSymbol)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("Binance returned a non 200 status code")
	}

	var bResponse BinanceResponse

	if err := json.NewDecoder(resp.Body).Decode(&bResponse); err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(bResponse.Price, 64)
	if err != nil {
		return 0, errors.New("Could not convert price")
	}

	return price, nil

}

