package btcwatcher

import (
	"encoding/json"
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

		// Parse the response
		var txs []Transaction
		if err := json.Unmarshal(resp.Body(), &txs); err != nil {
			log.Println("Error decoding JSON:", err)
			return
		}

		tx := txs[0] // Get the latest transaction

		// Print Transaction Details
		log.Println("Transaction ID:", tx.TxID)
		log.Println("Size:", tx.Size, "bytes")
		log.Println("Fee:", tx.Fee, "satoshis")
		log.Println("Confirmed:", tx.Status.Confirmed)
		log.Println("Block Height:", tx.Status.BlockHeight)

		// Print Inputs (vin)
		log.Println("\n🔹 Inputs (vin):")
		for _, vin := range tx.Vin {
			log.Printf("- From: %s (Spent %.8f BTC)\n", vin.Prevout.Address, float64(vin.Prevout.Value)/1e8)
		}

		// Print Outputs (vout)
		log.Println("\n🔹 Outputs (vout):")
		for _, vout := range tx.Vout {
			log.Printf("- To: %s (Received %.8f BTC)\n", vout.ScriptPubKeyAddress, float64(vout.Value)/1e8)
		}

		log.Println()

		time.Sleep(10 * time.Second) // Poll every 10 sec
	}
}
