package main

import (
	"log"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/watcher"
)

func main() {
	network := "mainnet"

	// addresses
	addresses := []string{
		"34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo",                             // Binance
		"3M219KR5vEneNb47ewrPfWyb5jQ2DjxRP6",                             // Binance
		"3PXBET2GrTwCamkeDzKCx8DeGDyrbuGKoc",                             // Binance
		"bc1qgdjqv0av3q56jvd82tkdjpy7gdp9ut8tlqmgrpmv24sq90ecnvqqjwvw97", // Bitfinex
		"bc1q4j7fcl8zx5yl56j00nkqez9zf3f6ggqchwzzcs5hjxwqhsgxvavq3qfgpr", // Coincheck
		"bc1qjasf9z3h7w3jspkhtgatgpyvvzgpa2wwd2lr0eh5tx44reyn2k7sfc27a4", // Tether
		"bc1p5fdr2ht0y4rjckn869skpml7pulm8wx6lu4c5eezwngx3c3uupzssx4myf",
	}

	blockConfirmedRequired := 1

	// Initialize the BTC Watcher
	btcWatcher := watcher.NewBTCWatcher(network, addresses, blockConfirmedRequired)

	// Start watching in a goroutine
	go btcWatcher.Run()
	defer btcWatcher.Close()

	// Receive new transactions from the channel
	for tx := range btcWatcher.OutputChannel {
		log.Println(*tx)
	}
}
