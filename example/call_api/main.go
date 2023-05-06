package main

import (
	"encoding/json"
	"os"

	"github.com/shayne/stock-ticker-stream-deck-plugin/pkg/api"
)

func main() {
	result := api.Call([]string{"AAPL"})
	bs, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(bs)
}
