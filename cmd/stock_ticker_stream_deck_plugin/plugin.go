package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shayne/go-streamdeck-sdk"
)

type tile struct {
	context string
	title   string
	symbol  string
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

func (p *plugin) renderTile(t *tile, data Result) *[]byte {
	var price, change, changePercent float64
	var status string
	statusColor := orange // regular/pre
	switch data.MarketState {
	case "REGULAR":
		status = ""
		price = data.RegularMarketPrice
		change = data.RegularMarketChange
		changePercent = data.RegularMarketChangePercent
	case "POST", "POSTPOST", "PREPRE", "CLOSED":
		statusColor = blue
		status = ""
		price = data.PostMarketPrice
		change = data.PostMarketChange
		changePercent = data.PostMarketChangePercent
	case "PRE":
		status = ""
		if data.PreMarketPrice > 0 {
			price = data.PreMarketPrice
			change = data.PreMarketChange
		} else {
			price = data.PostMarketPrice
			change = data.PostMarketChange
		}
		changePercent = data.PreMarketChangePercent
	}
	arrow := ""
	arrowColor := red
	if change > 0 {
		arrow = ""
		arrowColor = green
	} else if change == 0 {
		arrow = ""
	}
	title := data.Symbol
	if t.title != "" {
		title = t.title
	}
	return DrawTile(title, price, change, changePercent, status, statusColor, arrow, arrowColor)
}

func (p *plugin) updateTiles(tiles []*tile) {
	symbols := make([]string, 0, len(tiles))
	for _, t := range tiles {
		if t.symbol != "" {
			symbols = append(symbols, t.symbol)
		}
	}
	stocks := CallAPI(symbols)
	if stocks == nil {
		return
	}
	for _, t := range tiles {
		b := p.renderTile(t, stocks[t.symbol])
		err := p.sd.SetImage(t.context, *b)
		if err != nil {
			log.Fatalf("sd.SetImage: %v\n", err)
		}
	}
}

func (p *plugin) startUpdateLoop() {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			var tiles []*tile
			for _, t := range p.tiles {
				tiles = append(tiles, t)
			}
			p.updateTiles(tiles)
		}
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
		t := &tile{context: ev.Context, symbol: settings["symbol"]}
		p.tiles[ev.Context] = t
		p.updateTiles([]*tile{t})
	}
}

func (p plugin) OnTitleParametersDidChange(ev *streamdeck.EvTitleParametersDidChange) {
	t := p.tiles[ev.Context]
	if t == nil {
		log.Println("OnTitleParametersDidChange: Tile not found")
		return
	}
	t.title = ev.Payload.Title
	p.updateTiles([]*tile{t})
}

func (p plugin) OnPropertyInspectorConnected(ev *streamdeck.EvSendToPlugin) {
	if t, ok := p.tiles[ev.Context]; ok {
		settings := make(settingsType)
		settings["symbol"] = t.symbol
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
		switch sdpi.Key {
		case "symbol":
			symbol := strings.ToUpper(sdpi.Value)
			settings := make(settingsType)
			settings["symbol"] = symbol
			err = p.sd.SetSettings(ev.Context, &settings)
			if err != nil {
				log.Fatalf("setSettings: %v", err)
			}
			t := p.tiles[ev.Context]
			if t == nil {
				log.Fatal("Tile was nil")
			}
			t.symbol = symbol
			p.updateTiles([]*tile{t})
		}
	}
}

func (p plugin) OnApplicationDidLaunch(*streamdeck.EvApplication) {

}

func (p plugin) OnApplicationDidTerminate(*streamdeck.EvApplication) {

}
