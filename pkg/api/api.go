package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go/v2"
)

var (
	apiMu     sync.Mutex
	apiKey    string
	apiClient *finnhub.APIClient
)

func APIKey() string {
	apiMu.Lock()
	defer apiMu.Unlock()
	return apiKey
}

func SetAPIKey(key string) {
	apiMu.Lock()
	defer apiMu.Unlock()
	log.Printf("Setting API key to %q\n", key)
	if key == "" {
		clientMu.Lock()
		defer clientMu.Unlock()
	}
	apiKey = key
}

var (
	clientMu sync.Mutex
)

func APIService() *finnhub.DefaultApiService {
	clientMu.Lock()
	defer clientMu.Unlock()

	apiKey := APIKey()
	if apiKey == "" {
		return nil
	}
	config := finnhub.NewConfiguration()
	config.AddDefaultHeader("X-Finnhub-Token", apiKey)
	apiClient = finnhub.NewAPIClient(config)
	return apiClient.DefaultApi
}

type APIResult struct {
	Quotes       map[string]finnhub.Quote
	MarketStatus finnhub.MarketStatus
}

var (
	marketStatus     *finnhub.MarketStatus
	marketStatusTime time.Time
	lastCallMu       sync.Mutex
	lastCallTime     time.Time
)

// Call makes HTTP request for stock data
func Call(symbols []string) *APIResult {
	if len(symbols) == 0 {
		return nil
	}

	lastCallMu.Lock()
	timeSinceLast := time.Since(lastCallTime)
	if timeSinceLast < 1*time.Second {
		time.Sleep(1*time.Second - timeSinceLast)
	}
	lastCallMu.Unlock()

	defer func() {
		lastCallMu.Lock()
		lastCallTime = time.Now()
		lastCallMu.Unlock()
	}()

	service := APIService()
	if service == nil {
		log.Printf("API client was nil\n")
		return nil
	}

	ctx := context.Background()

	if marketStatus == nil || time.Since(marketStatusTime) >= 5*time.Minute {
		for {
			ms, response, err := service.MarketStatus(ctx).Exchange("US").Execute()
			marketStatus = &ms
			if err != nil {
				if response != nil && response.StatusCode == 429 {
					log.Printf("Rate limit exceeded, waiting 5 seconds\n")
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("Error getting market status: %v", err)
				return nil
			}
			break
		}
		marketStatusTime = time.Now()

		// Rate limit is 60 requests per minute
		time.Sleep(1 * time.Second)
	}

	quotes := make(map[string]finnhub.Quote)
	for i, symbol := range symbols {
		quote, err := getQuote(ctx, symbol)
		if err != nil {
			log.Printf("Error getting quote for %s: %v\n", symbol, err)
		} else {
			quotes[symbol] = *quote
		}
		if len(symbols) > 1 && i < len(symbols)-1 {
			// Rate limit is 60 requests per minute
			time.Sleep(1 * time.Second)
		}
	}
	return &APIResult{
		Quotes:       quotes,
		MarketStatus: *marketStatus,
	}
}

func getQuote(ctx context.Context, symbol string) (quote *finnhub.Quote, err error) {
	service := APIService()
	log.Printf("Getting quote for %s ", symbol)
	for {
		var q finnhub.Quote
		var response *http.Response
		q, response, err = service.Quote(ctx).Symbol(symbol).Execute()
		if err != nil {
			if response != nil && response.StatusCode == 429 {
				log.Printf("Rate limit exceeded, waiting 5 seconds\n")
				time.Sleep(5 * time.Second)
				continue
			}
		}
		quote = &q
		break
	}
	if err != nil {
		return nil, fmt.Errorf("error getting quote for %s: %v", symbol, err)
	}
	return quote, nil
}
