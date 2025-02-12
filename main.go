package main

import (
	"log"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/watcher"
)

func main() {
	network := "mainnet"

	addresses := []string{
		"1LdRcdxfbSnmCYYNdeYpUnztiYzVfBEQeC",
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
