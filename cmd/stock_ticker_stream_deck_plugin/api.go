package main

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const apiURL = "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=symbol,marketState,regularMarketPrice,regularMarketChange,regularMarketChangePercent,preMarketPrice,preMarketChange,preMarketChangePercent,postMarketPrice,postMarketChange,postMarketChangePercent"

// StockData contains price data for stock symbol
type StockData struct {
	Symbol            string  `json:"symbol"`
	Price             float64 `json:"regularMarketPrice"`
	Change            float64 `json:"regularMarketChange"`
	ChangePercent     float64 `json:"regularMarketChangePercent"`
	PostPrice         float64 `json:"postMarketPrice"`
	PostChange        float64 `json:"postMarketChange"`
	PostChangePercent float64 `json:"postMarketChangePercent"`
	PrePrice          float64 `json:"preMarketPrice"`
	PreChange         float64 `json:"preMarketChange"`
	PreChangePercent  float64 `json:"preMarketChangePercent"`
	MarketState       string  `json:"marketState"`
}

type apiResult struct {
	Result []StockData `json:"result"`
}

// CallAPI makes HTTP request for stock data
func CallAPI(symbols []string) map[string]StockData {
	client := &http.Client{}

	url := apiURL + "&symbols=" + strings.Join(symbols, ",")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("accept-encoding", "gzip, deflate, br")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("dnt", "1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.90 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	reader, err := gzip.NewReader(resp.Body)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalln(err)
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		log.Fatal("body unmarshal", err)
	}
	var results apiResult
	err = json.Unmarshal(*objmap["quoteResponse"], &results)
	if err != nil {
		log.Fatal("results unmarshal", err)
	}

	ret := make(map[string]StockData)
	for _, v := range results.Result {
		ret[v.Symbol] = v
	}

	return ret
}
