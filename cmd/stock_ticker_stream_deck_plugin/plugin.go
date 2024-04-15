package main

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go/v2"
	"github.com/gorilla/websocket"
	"github.com/shayne/go-streamdeck-sdk"

	"github.com/shayne/stock-ticker-stream-deck-plugin/pkg/api"
)

type tile struct {
	context string
	title   string
	symbol  string
	apikey  string
}

type plugin struct {
	sd    *streamdeck.StreamDeck
	tiles map[string]*tile
}

type evSdpiCollection struct {
	Group     bool     `json:"group"`
	Index     int      `json:"index"`
	Key       string   `json:"key"`
	Selection []string `json:"selection"`
	Value     string   `json:"value"`
}

type settingsType map[string]string

func newPlugin(port, uuid, event, info string) *plugin {
	sd := streamdeck.NewStreamDeck(port, uuid, event, info)
	p := &plugin{sd: sd, tiles: make(map[string]*tile)}
	sd.SetDelegate(p)
	return p
}

func (p *plugin) renderTile(t *tile, symbol string, marketStatus finnhub.MarketStatus, quote finnhub.Quote) *[]byte {
	var price, change, changePercent float32
	var status string
	statusColor := orange // regular/pre

	switch marketStatus.GetSession() {
	case "regular":
		status = ""
	case "pre-market":
		status = ""
	default:
		statusColor = blue
		status = ""
	}
	price = quote.GetC()
	change = quote.GetD()
	changePercent = quote.GetDp()
	arrow := ""
	arrowColor := red
	if change > 0 {
		arrow = ""
		arrowColor = green
	} else if change == 0 {
		arrow = ""
	}
	title := symbol
	if t.title != "" {
		title = t.title
	}
	return DrawTile(title, price, change, changePercent, status, statusColor, arrow, arrowColor)
}

var (
	updateMu = sync.Mutex{}
)

func (p *plugin) updateTiles(tiles []*tile) {
	go p.goUpdateTiles(tiles)
}

func (p *plugin) goUpdateTiles(tiles []*tile) {
	updateMu.Lock()
	defer updateMu.Unlock()

	var symbols []string
	for _, t := range tiles {
		if t.symbol != "" {
			symbols = append(symbols, t.symbol)
		}
	}
	if len(symbols) == 0 {
		return
	}
	result := api.Call(symbols)
	if result == nil {
		log.Print("API call returned nil")
		return
	}
	for _, t := range tiles {
		quote, ok := result.Quotes[t.symbol]
		if !ok {
			continue
		}
		b := p.renderTile(t, t.symbol, result.MarketStatus, quote)
		err := p.sd.SetImage(t.context, *b)
		if err != nil {
			log.Fatalf("sd.SetImage: %v\n", err)
		}
	}
}

func (p *plugin) startUpdateLoop() {
	tick := time.Tick(5 * time.Minute)
	for range tick {
		p.updateAllTiles()
	}
}

func (p *plugin) Run() {
	err := p.sd.Connect()
	if err != nil {
		log.Fatalf("Connect: %v\n", err)
	}
	go p.startUpdateLoop()
	p.sd.ListenAndWait()
}

func (p plugin) OnConnected(*websocket.Conn) {
}

func (p plugin) OnWillAppear(ev *streamdeck.EvWillAppear) {
	if t, ok := p.tiles[ev.Context]; ok {
		p.updateTiles([]*tile{t})
	} else {
		var settings settingsType
		err := json.Unmarshal(*ev.Payload.Settings, &settings)
		if err != nil {
			log.Println("OnWillAppear settings unmarshal", err)
		}
		// If the API key was not set, but these settings
		// have an API key, set it and update all tiles
		var updateAll bool
		apiKey := api.APIKey()
		if apiKey == "" && settings["apikey"] != "" {
			apiKey = settings["apikey"]
			api.SetAPIKey(apiKey)
			updateAll = true
		}
		t := &tile{context: ev.Context, symbol: settings["symbol"], apikey: apiKey}
		p.tiles[ev.Context] = t
		if updateAll {
			for _, t := range p.tiles {
				t.apikey = apiKey
			}
			p.updateAllTiles()
		} else {
			p.updateTiles([]*tile{t})
		}
	}
}

func (p plugin) updateAllTiles() {
	var tiles []*tile
	for _, t := range p.tiles {
		tiles = append(tiles, t)
	}
	p.updateTiles(tiles)
}

func (p plugin) OnTitleParametersDidChange(ev *streamdeck.EvTitleParametersDidChange) {
	t := p.tiles[ev.Context]
	if t == nil {
		log.Println("OnTitleParametersDidChange: Tile not found")
		return
	}
	t.title = ev.Payload.Title
}

func (p plugin) OnPropertyInspectorConnected(ev *streamdeck.EvSendToPlugin) {
	if t, ok := p.tiles[ev.Context]; ok {
		settings := make(settingsType)
		settings["symbol"] = t.symbol
		settings["apikey"] = api.APIKey()
		p.sd.SendToPropertyInspector(ev.Action, ev.Context, &settings)
	}
}

func (p plugin) OnSendToPlugin(ev *streamdeck.EvSendToPlugin) {
	var payload map[string]*json.RawMessage
	err := json.Unmarshal(*ev.Payload, &payload)
	if err != nil {
		log.Println("OnSendToPlugin unmarshal", err)
	}
	if data, ok := payload["sdpi_collection"]; ok {
		sdpi := evSdpiCollection{}
		err = json.Unmarshal(*data, &sdpi)
		if err != nil {
			log.Println("SDPI unmarshal", err)
		}
		t := p.tiles[ev.Context]
		if t == nil {
			log.Printf("Tile was nil, creating new tile for %s\n", ev.Context)
			p.tiles[ev.Context] = &tile{context: ev.Context}
			t = p.tiles[ev.Context]
		}
		settings := make(settingsType)
		settings["apikey"] = api.APIKey()
		settings["symbol"] = p.tiles[ev.Context].symbol

		var updateAll bool
		switch sdpi.Key {
		case "symbol":
			symbol := strings.ToUpper(sdpi.Value)
			t.symbol = symbol
			settings["symbol"] = symbol
		case "apikey":
			// If they updated the API key, we need to update it
			// and update all tiles
			apikey := sdpi.Value
			api.SetAPIKey(apikey)
			updateAll = true
			t.apikey = apikey
			settings["apikey"] = apikey
		}
		if updateAll {
			apikey := api.APIKey()
			for _, t := range p.tiles {
				t.apikey = apikey
				err := p.sd.SetSettings(t.context, settingsType{
					"symbol": t.symbol,
					"apikey": apikey,
				})
				if err != nil {
					log.Printf("error setting settings: %v", err)
				}
			}
			p.updateAllTiles()
		} else {
			log.Printf("saving settings: %v", settings)
			err = p.sd.SetSettings(ev.Context, &settings)
			if err != nil {
				log.Fatalf("setSettings: %v", err)
			}
			p.updateTiles([]*tile{t})
		}
	}
}

func (p plugin) OnApplicationDidLaunch(*streamdeck.EvApplication) {

}

func (p plugin) OnApplicationDidTerminate(*streamdeck.EvApplication) {

}
