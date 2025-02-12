package watcher

import (
	"encoding/json"
	"fmt"
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
func (w *BTCWatcher) getNewTransactions() error {
	// Get the latest blocks at the current height + 1
	blocks := w.getBlocks(w.LastBlockHeight + 1)
	if blocks == nil {
		return fmt.Errorf("error fetching blocks")
	}

	// Update the last block height
	latestBlockHeight := blocks[len(blocks)-1].Height
	w.updateLastBlockHeight(latestBlockHeight)

	// Get transactions from the latest blocks
	var transactions []*Transaction
	for _, block := range blocks {
		txs := w.getTransactionsFromBlock(block)
		transactions = append(transactions, txs...)
	}

	return nil
}

// getBlock gets the block at the given height
func (w *BTCWatcher) getBlocks(start_height int) []*Block {
	// Get the block at the given height
	resp, err := w.Client.R().Get(fmt.Sprintf("%s/blocks/%d", w.BaseUrl, start_height))
	if err != nil {
		return nil
	}

	// Parse the response
	var blocks []*Block
	if err := json.Unmarshal(resp.Body(), &blocks); err != nil {
		return nil
	}

	return blocks
}

// updateLastBlockHeight updates the last block height
func (w *BTCWatcher) updateLastBlockHeight(height int) {
	w.LastBlockHeight = height
}

// getTransactionsFromBlock gets transactions from the given block
func (w *BTCWatcher) getTransactionsFromBlock(block *Block) []*Transaction {
	var transactions []*Transaction
	txCount := block.TxCount
	for i := 0; i < txCount; i += 25 {
		// Get transactions for the block
		resp, err := w.Client.R().Get(fmt.Sprintf("%s/block/%s", w.BaseUrl, block.ID))
		if err != nil {
			return nil
		}

		// Parse the response
		var txs []*Transaction
		if err := json.Unmarshal(resp.Body(), &txs); err != nil {
			return nil
		}

		transactions = append(transactions, txs...)
	}

	return transactions
}
