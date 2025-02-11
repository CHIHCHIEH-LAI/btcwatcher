package main

import "github.com/CHIHCHIEH-LAI/btcwatcher/btcwatcher"

func main() {
	cfg := &btcwatcher.Config{
		Network: "mainnet", // Use "mainnet" for real BTC
		Address: "bc1qd4ysezhmypwty5dnw7c8nqy5h5nxg0xqsvaefd0qn5kq32vwnwqqgv4rzr",
	}

	watcher := btcwatcher.NewWatcher(cfg)
	watcher.WatchTransactions()
}
