package main

import (
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/watcher"
)

func main() {
	network := "mainnet"

	addresses := []string{
		"tb1qexampleaddress1",
		"tb1qexampleaddress2",
	}

	// Initialize the BTC Watcher
	btcWatcher := watcher.NewBTCWatcher(network, addresses)

	// Start watching in a goroutine
	go btcWatcher.Run()
	defer btcWatcher.Close()

	// Receive new transactions from the channel
	// for tx := range btcWatcher.TxChannel {
	// 	fmt.Printf("✅ New Transaction: %s -> %s (%.8f BTC)\n", tx.TxID, tx.Address, float64(tx.Value)/1e8)
	// }
}
