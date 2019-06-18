plugin:
	-@kill-streamdeck.bat
	@go build -o com.exension.stockticker.sdPlugin\\sdplugin-stockticker.exe github.com/exension/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@xcopy com.exension.stockticker.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.stockticker.sdPlugin\\ /E /Q /Y
	@start-streamdeck.bat

debug:
	@go build -o com.exension.stockticker.sdPlugin\\sdplugin-stockticker.exe github.com/exension/stock-ticker-stream-deck-plugin/cmd/stock_ticker_stream_deck_plugin
	@xcopy com.exension.stockticker.sdPlugin $(APPDATA)\\Elgato\\StreamDeck\\Plugins\\com.exension.stockticker.sdPlugin\\ /E /Q /Y

release:
	-@rm release\com.exension.stockticker.streamDeckPlugin
	@DistributionTool.exe com.exension.stockticker.sdPlugin release

.PHONY: release