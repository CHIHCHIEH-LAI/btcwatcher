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
			w.watchNewTxsFromWatchedAddresses()
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

// watchNewTxsFromWatchedAddresses watches new transactions from the watched addresses
func (w *BTCWatcher) watchNewTxsFromWatchedAddresses() {
	// Get new blocks
	blocks := w.getNewBlocks()
	if blocks == nil {
		return
	}

	// Get transactions from the blocks
	txs := w.getTxsFromBlocks(blocks)

	// Filter transactions by watched addresses
	filteredTxs := w.filterTxsByWatchedAddresses(txs)

	// Send transactions to the channel
	w.sendTransactionsToChannel(filteredTxs)
}

// getNewBlocks gets new blocks
func (w *BTCWatcher) getNewBlocks() []*Block {
	blocks := w.getBlocks(w.LastBlockHeight + 1)

	// Return if no new blocks
	if blocks == nil {
		return nil
	}

	// Update the last block height
	latestBlockHeight := blocks[len(blocks)-1].Height
	w.updateLastBlockHeight(latestBlockHeight)

	return blocks
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

// getTxsFromBlocks gets transactions from the given blocks
func (w *BTCWatcher) getTxsFromBlocks(blocks []*Block) []*Transaction {
	var transactions []*Transaction
	for _, block := range blocks {
		transactions = append(transactions, w.getTxsFromBlock(block)...)
	}

	return transactions
}

// getTxsFromBlock gets transactions from the given block
func (w *BTCWatcher) getTxsFromBlock(block *Block) []*Transaction {
	var transactions []*Transaction
	txCount := block.TxCount
	for i := 0; i < txCount+25; i += 25 {
		// Get transactions for the block
		resp, err := w.Client.R().Get(fmt.Sprintf("%s/block/%s/txs", w.BaseUrl, block.ID))
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

// filterTxsByWatchedAddresses filters transactions by watched addresses
func (w *BTCWatcher) filterTxsByWatchedAddresses(txs []*Transaction) []*Transaction {
	var filteredTxs []*Transaction
	for _, tx := range txs {
		if w.txContainsWatchedAddress(tx) {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	return filteredTxs
}

// txContainsWatchedAddress checks if the transaction contains a watched address
func (w *BTCWatcher) txContainsWatchedAddress(tx *Transaction) bool {
	for _, vout := range tx.Vout {
		if w.WatchedAddresses[vout.ScriptPubKeyAddress] {
			return true
		}
	}

	return false
}

// sendTransactionsToChannel sends transactions to the channel
func (w *BTCWatcher) sendTransactionsToChannel(txs []*Transaction) {
	for _, tx := range txs {
		w.TxChannel <- *tx
	}
}
