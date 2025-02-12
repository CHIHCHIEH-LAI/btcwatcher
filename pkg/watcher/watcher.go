package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type BTCWatcher struct {
	Client           *resty.Client
	BaseUrl          string
	WatchedAddresses map[string]bool
	LastBlockHeight  int
	TxChannel        chan Transaction
	StopRunning      chan bool
}

// NewBTCWatcher creates a new BTCWatcher instance with the given network and watched addresses
func NewBTCWatcher(network string, watchedAddresses []string) *BTCWatcher {
	btcwatcher := &BTCWatcher{
		Client:          resty.New(),
		LastBlockHeight: 0,
		TxChannel:       make(chan Transaction),
		StopRunning:     make(chan bool),
	}

	btcwatcher.setWatchedAddresses(watchedAddresses)
	btcwatcher.setBaseUrl(network)

	return btcwatcher
}

// setWatchedAddresses sets the addresses to watch
func (w *BTCWatcher) setWatchedAddresses(addresses []string) {
	w.WatchedAddresses = make(map[string]bool)
	for _, addr := range addresses {
		w.WatchedAddresses[addr] = true
	}
}

// setBaseUrl sets the base URL for the Blockstream API
func (w *BTCWatcher) setBaseUrl(network string) {
	baseURL := "https://blockstream.info/api"
	if network == "testnet" {
		baseURL = "https://blockstream.info/testnet/api"
	}

	w.BaseUrl = baseURL
}

// Run starts monitoring new blocks
func (w *BTCWatcher) Run() {
	for {
		select {
		case <-w.StopRunning:
			return
		default:
			w.getNewTransactions()
			time.Sleep(60 * time.Second) // Check every 10 sec
		}
	}
}

// Close closes the btcwatcher
func (w *BTCWatcher) Close() {
	w.StopRunning <- true
	close(w.TxChannel)
	close(w.StopRunning)
}

// getNewTransactions gets new transactions from the Blockstream API
func (w *BTCWatcher) getNewTransactions() {
	latestBlockHeight, err := w.getLatestBlockHeight()
	log.Println("Latest block height:", latestBlockHeight)
	if err != nil {
		log.Println("Error fetching latest block height:", err)
		return
	}
	if latestBlockHeight > w.LastBlockHeight {
		// Process missing blocks
		// for height := w.LastBlockHeight + 1; height <= latestBlockHeight; height++ {
		// 	block := w.getBlock(height)
		// 	if block == nil {
		// 		continue
		// 	}
		// 	log.Println("Processing block:", block.ID)
		// }
		w.LastBlockHeight = latestBlockHeight
	}
}

// getLatestBlockHeight gets the latest block height
func (w *BTCWatcher) getLatestBlockHeight() (int, error) {
	resp, err := w.Client.R().Get(fmt.Sprintf("%s/blocks/tip/height", w.BaseUrl))
	if err != nil {
		return 0, fmt.Errorf("error fetching latest block height: %v", err)
	}

	var height int
	if err := json.Unmarshal(resp.Body(), &height); err != nil {
		return 0, fmt.Errorf("error decoding JSON: %v", err)
	}

	return height, nil
}

// getBlock gets the block at the given height
func (w *BTCWatcher) getBlock(height int) *Block {
	resp, err := w.Client.R().Get(fmt.Sprintf("%s/block/%d", w.BaseUrl, height))
	if err != nil {
		log.Println("Error fetching block:", err)
		return nil
	}

	var block Block
	if err := json.Unmarshal(resp.Body(), &block); err != nil {
		log.Println("Error decoding JSON:", err)
		return nil
	}

	return &block
}

// // WatchTransactions polls the Blockstream API for new transactions
// func (w *BTCWatcher) WatchTransactions() {
// 	log.Println("Watching BTC address:", w.Config.Address)

// 	for {
// 		// Get transactions for the address
// 		resp, err := w.Client.R().Get(fmt.Sprintf("/address/%s/txs", w.Config.Address))
// 		if err != nil {
// 			log.Println("Error fetching transactions:", err)
// 			time.Sleep(10 * time.Second)
// 			continue
// 		}

// 		// Parse the response
// 		var txs []Transaction
// 		if err := json.Unmarshal(resp.Body(), &txs); err != nil {
// 			log.Println("Error decoding JSON:", err)
// 			return
// 		}

// 		// Get the number of transactions
// 		log.Println("Found", len(txs), "transactions")

// 		// Get the latest transaction
// 		tx := txs[0]

// 		// Print Transaction Details
// 		log.Println("Transaction ID:", tx.TxID)
// 		log.Println("Size:", tx.Size, "bytes")
// 		log.Println("Fee:", tx.Fee, "satoshis")
// 		log.Println("Confirmed:", tx.Status.Confirmed)
// 		log.Println("Block Height:", tx.Status.BlockHeight)

// 		// Print Inputs (vin)
// 		log.Println("\n🔹 Inputs (vin):")
// 		for _, vin := range tx.Vin {
// 			log.Printf("- From: %s (Spent %.8f BTC)\n", vin.Prevout.Address, float64(vin.Prevout.Value)/1e8)
// 		}

// 		// Print Outputs (vout)
// 		log.Println("\n🔹 Outputs (vout):")
// 		for _, vout := range tx.Vout {
// 			log.Printf("- To: %s (Received %.8f BTC)\n", vout.ScriptPubKeyAddress, float64(vout.Value)/1e8)
// 		}

// 		log.Println()

// 		time.Sleep(10 * time.Second) // Poll every 10 sec
// 	}
// }
