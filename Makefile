plugin:
	-@kill-streamdeck.bat
	@go build -o com.shayne.stocks.sdPlugin\\sdplugin-stocks.exe github.com/shayne/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@env GOOS=darwin GOARCH=amd64 go build -o com.shayne.stocks.sdPlugin\\sdplugin-stocks github.com/shayne/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@xcopy com.shayne.stocks.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.shayne.stocks.sdPlugin\\ /E /Q /Y
	@start-streamdeck.bat

debug:
	@go build -o com.shayne.stocks.sdPlugin\\sdplugin-stocks.exe github.com/shayne/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@env GOOS=darwin GOARCH=amd64 go build -o com.shayne.stocks.sdPlugin\\sdplugin-stocks github.com/shayne/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@xcopy com.shayne.stocks.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.shayne.stocks.sdPlugin\\ /E /Q /Y

release:
	-@rm release\com.shayne.stocks.streamDeckPlugin
	@DistributionTool.exe com.shayne.stocks.sdPlugin release

.PHONY: release