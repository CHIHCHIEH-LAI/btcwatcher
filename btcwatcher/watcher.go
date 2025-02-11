package btcwatcher

import (
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type Watcher struct {
	Config *Config
	Client *resty.Client
}

// NewWatcher initializes the watcher
func NewWatcher(cfg *Config) *Watcher {
	baseURL := "https://blockstream.info/api"
	if cfg.Network == "testnet" {
		baseURL = "https://blockstream.info/testnet/api"
	}

	return &Watcher{
		Config: cfg,
		Client: resty.New().SetBaseURL(baseURL),
	}
}

// WatchTransactions polls the Blockstream API for new transactions
func (w *Watcher) WatchTransactions() {
	log.Println("Watching BTC address:", w.Config.Address)

	for {
		// Get transactions for the address
		resp, err := w.Client.R().Get(fmt.Sprintf("/address/%s/txs", w.Config.Address))
		if err != nil {
			log.Println("Error fetching transactions:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("Latest Transactions:", resp.String())

		time.Sleep(10 * time.Second) // Poll every 10 sec
	}
}
