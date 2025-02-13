package main

import (
	"encoding/json"
	"log"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/watcher"
)

func main() {
	network := "mainnet"

	fromHeight := 880000

	// addresses
	addresses := []string{
		// "34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo",                             // Binance
		// "3M219KR5vEneNb47ewrPfWyb5jQ2DjxRP6",                             // Binance
		// "3PXBET2GrTwCamkeDzKCx8DeGDyrbuGKoc",                             // Binance
		// "bc1qgdjqv0av3q56jvd82tkdjpy7gdp9ut8tlqmgrpmv24sq90ecnvqqjwvw97", // Bitfinex
		// "bc1q4j7fcl8zx5yl56j00nkqez9zf3f6ggqchwzzcs5hjxwqhsgxvavq3qfgpr", // Coincheck
		// "bc1qjasf9z3h7w3jspkhtgatgpyvvzgpa2wwd2lr0eh5tx44reyn2k7sfc27a4", // Tether
		"1K6KoYC69NnafWJ7YgtrpwJxBLiijWqwa6",         // Block 883,596
		"bc1qs3w5wma79pp87yneq9gj6x6exf8snksud8lgr3", // Block 883,591
		"bc1q3vgzshdq6dg4h5e76xjr59rrj3tfyv2uamkxcj", // Block 883,584
	}

	blockConfirmedRequired := 1

	// Initialize the BTC Watcher
	btcWatcher := watcher.NewBTCWatcher(network, fromHeight, addresses, blockConfirmedRequired)

	// Start watching in a goroutine
	go btcWatcher.Run()
	defer btcWatcher.Close()

	// Receive new transactions from the channel
	for tx := range btcWatcher.OutputChannel {
		txData, err := json.MarshalIndent(*tx, "", "  ")
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(string(txData))
	}
}
