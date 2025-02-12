package main

import (
	"log"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/watcher"
)

func main() {
	network := "mainnet"

	addresses := []string{
		"34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo",                             // Binance
		"bc1qgdjqv0av3q56jvd82tkdjpy7gdp9ut8tlqmgrpmv24sq90ecnvqqjwvw97", // Bitfinex
	}

	// Initialize the BTC Watcher
	btcWatcher := watcher.NewBTCWatcher(network, addresses)

	// Start watching in a goroutine
	go btcWatcher.Run()
	defer btcWatcher.Close()

	// Receive new transactions from the channel
	for tx := range btcWatcher.TxChannel {
		log.Println(tx)
	}
}
