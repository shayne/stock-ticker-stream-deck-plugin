plugin:
	-@./kill-streamdeck.sh
	@env GOOS=windows GOARCH=amd64 go build -o com.exension.stocks.sdPlugin/sdplugin-stocks.exe ./...
	@env GOOS=darwin GOARCH=amd64 go build -o com.exension.stocks.sdPlugin/sdplugin-stocks ./...
	@cp -r com.exension.stocks.sdPlugin ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/
	@./start-streamdeck.sh

debug:
	@env GOOS=windows GOARCH=amd64 go build com.exension.stocks.sdPlugin/sdplugin-stocks.exe github.com/exension/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@env GOOS=darwin GOARCH=amd64 go build -o com.exension.stocks.sdPlugin/sdplugin-stocks github.com/exension/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@cp -r com.exension.stocks.sdPlugin ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/com.exension.stocks.sdPlugin

release:
	@env GOOS=windows GOARCH=amd64 go build -o com.exension.stocks.sdPlugin/sdplugin-stocks.exe ./...
	@env GOOS=darwin GOARCH=amd64 go build -o com.exension.stocks.sdPlugin/sdplugin-stocks ./...
	-@rm release\com.exension.stocks.streamDeckPlugin
	@./DistributionTool -b -i com.exension.stocks.sdPlugin -o release

.PHONY: release