package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Exchange interface {
	GetPrice(pair string) (float64, error)
	GetName() string
}

type Graph struct {
	Rates map[string]map[string]float64 // map of a map - first string is base currency, second is converted currency, float is exchange rate
	mu    sync.Mutex
}

type Coinbase struct {
}

type ArbitrageResult struct {
	Path   []string
	Profit float64
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

func main() {

	coinbase := Coinbase{}

	ticker := time.NewTicker(5 * time.Second)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	fmt.Printf("Bot started, Ctrl + C to stop\n")

	for {
		select {
		case <-ticker.C:

			market := NewGraph()
			fee := 0.996

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				btcBasePrice, err1 := coinbase.GetPrice("BTC-USD")
				if err1 == nil {
					market.AddRate("USD", "BTC", (1/btcBasePrice)*fee)
					market.AddRate("BTC", "USD", btcBasePrice*fee)
				}
			}()

			altCoins := []string{"ETH", "SOL", "XRP", "DOGE", "ADA", "LINK"}

			for _, coin := range altCoins {

				wg.Add(1)

				go func(c string) {
					defer wg.Done()
					usdPair := fmt.Sprintf("%s-USD", c)
					btcPair := fmt.Sprintf("%s-BTC", c)

					usdPrice, errUSD := coinbase.GetPrice(usdPair)
					btcPrice, errBTC := coinbase.GetPrice(btcPair)

					if errUSD == nil && errBTC == nil {
						market.AddRate("USD", c, (1/usdPrice)*fee)
						market.AddRate(c, "USD", usdPrice*fee)

						market.AddRate("BTC", c, (1/btcPrice)*fee)
						market.AddRate(c, "BTC", btcPrice*fee)
					}

				}(coin)

			}

			wg.Wait()

			CalculateDynamicPath(market, 100.0, "USD", 100.0, []string{"USD"}, 4)

		case <-quit:
			fmt.Println("\n Quitting")
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

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Rates[source] == nil {

		g.Rates[source] = make(map[string]float64)
	}

	g.Rates[source][destination] = rate

}

func CalculateDynamicPath(g *Graph, startingamount float64, currentNode string, currentAmount float64, path []string, maxSteps int) []ArbitrageResult {
	res := []ArbitrageResult{}
	if len(path) > maxSteps {
		return res //exceeded max steps
	}

	if len(path) > 1 && path[0] == currentNode {
		if currentAmount > startingamount+0.02 {
			res = append(res, ArbitrageResult{
				Path:   append([]string{}, path...),
				Profit: currentAmount - startingamount,
			})
		}
		return res
	}

	for nextNode, rate := range g.Rates[currentNode] {

		newAmount := currentAmount * rate
		newPath := append([]string{}, path...)
		newPath = append(newPath, nextNode)

		results := CalculateDynamicPath(g, startingamount, nextNode, newAmount, newPath, maxSteps)
		res = append(res, results...)
	}
	return res
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
