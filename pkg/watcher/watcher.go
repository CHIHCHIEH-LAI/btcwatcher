package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type BTCWatcher struct {
	client            *resty.Client
	baseUrl           string
	watchedAddresses  map[string]bool
	lastBlockHeight   int
	blockChannel      chan *Block
	txChannel         chan *Transaction
	filteredTxChannel chan *Transaction
	OutputChannel     chan *Transaction
	stopRunning       chan bool
}

// NewBTCWatcher creates a new BTCWatcher instance with the given network and watched addresses
func NewBTCWatcher(network string, watchedAddresses []string) *BTCWatcher {
	btcwatcher := &BTCWatcher{
		client:            resty.New(),
		blockChannel:      make(chan *Block),
		txChannel:         make(chan *Transaction),
		filteredTxChannel: make(chan *Transaction),
		OutputChannel:     make(chan *Transaction),
		stopRunning:       make(chan bool),
	}

	btcwatcher.setBaseUrl(network)
	btcwatcher.setWatchedAddresses(watchedAddresses)

	btcwatcher.lastBlockHeight = btcwatcher.getLatestBlockHeight() - 1

	return btcwatcher
}

// setWatchedAddresses sets the addresses to watch
func (w *BTCWatcher) setWatchedAddresses(addresses []string) {
	w.watchedAddresses = make(map[string]bool)
	for _, addr := range addresses {
		w.watchedAddresses[addr] = true
	}
}

// setBaseUrl sets the base URL for the Blockstream API
func (w *BTCWatcher) setBaseUrl(network string) {
	baseURL := "https://blockstream.info/api"
	if network == "testnet" {
		baseURL = "https://blockstream.info/testnet/api"
	}

	w.baseUrl = baseURL
}

// Run starts monitoring new blocks
func (w *BTCWatcher) Run() {
	go w.watchForNewBlock()
	go w.collectTransactionsFromBlock()
	go w.filterTransactionByWatchedAddresses()
	go w.outputTransaction()
}

// collectNewBlock collects new block
func (w *BTCWatcher) watchForNewBlock() {
	for {
		select {
		case <-w.stopRunning:
			return
		default:
			w.fetchNewBlocks()
			time.Sleep(60 * time.Second)
		}
	}
}

// fetchNewBlocks fetches new blocks
func (w *BTCWatcher) fetchNewBlocks() {
	// Get the latest block height
	latestBlockHeight := w.getLatestBlockHeight()

	if latestBlockHeight <= w.lastBlockHeight {
		return
	}

	blocks := w.fetchBlocks(w.lastBlockHeight+1, latestBlockHeight)

	log.Printf("Fetched %d blocks from %d to %d", len(blocks), w.lastBlockHeight+1, latestBlockHeight)

	// Send blocks to the channel
	for _, block := range blocks {
		w.blockChannel <- block
	}

	// Update the last block height
	w.updateLastBlockHeight(latestBlockHeight)
}

// fetchBlocks fetches blocks from the given start height to the end height
func (w *BTCWatcher) fetchBlocks(start_height, end_height int) []*Block {
	var blocks []*Block
	for i := start_height; i <= end_height; i += 10 {
		// Get the blocks starting from the given height
		resp, err := w.client.R().Get(fmt.Sprintf("%s/blocks/%d", w.baseUrl, start_height))
		if err != nil {
			continue
		}

		// Parse the response
		var subBlocks []*Block
		if err := json.Unmarshal(resp.Body(), &subBlocks); err != nil {
			continue
		}

		blocks = append(blocks, subBlocks...)
	}

	// Trim the blocks
	heightDiff := end_height - start_height
	if len(blocks) > heightDiff {
		blocks = blocks[:heightDiff]
	}

	return blocks
}

// updateLastBlockHeight updates the last block height
func (w *BTCWatcher) updateLastBlockHeight(height int) {
	w.lastBlockHeight = height
}

// collectTransactionsFromBlock collects transactions from the block
func (w *BTCWatcher) collectTransactionsFromBlock() {
	for {
		select {
		case <-w.stopRunning:
			return
		case block := <-w.blockChannel:
			w.fetchTransactionsFromBlock(block)
		}
	}
}

// fetchTransactionsFromBlock fetches transactions from the block
func (w *BTCWatcher) fetchTransactionsFromBlock(block *Block) {
	txCount := block.TxCount
	for i := 0; i < txCount; i += 25 {
		// Fetch 25 transactions beginning at index i
		resp, err := w.client.R().Get(fmt.Sprintf("%s/block/%s/txs/%d", w.baseUrl, block.ID, i))
		if err != nil {
			continue
		}

		// Parse the response
		var txs []*Transaction
		if err := json.Unmarshal(resp.Body(), &txs); err != nil {
			continue
		}

		// Send transactions to the channel
		for _, tx := range txs {
			w.txChannel <- tx
		}
	}

	log.Printf("Fetched %d transactions from block %s", txCount, block.ID)
}

// filterTransactionByWatchedAddresses filters transaction by watched addresses
func (w *BTCWatcher) filterTransactionByWatchedAddresses() {
	for {
		select {
		case <-w.stopRunning:
			return
		case tx := <-w.txChannel:
			for _, vout := range tx.Vout {
				if w.watchedAddresses[vout.ScriptPubKeyAddress] {
					w.filteredTxChannel <- tx
					log.Printf("Transaction %s is sent to the filteredTxChannel", tx.TxID)
					break
				}
			}
		}
	}
}

// outputTransaction outputs transaction
func (w *BTCWatcher) outputTransaction() {
	for {
		select {
		case <-w.stopRunning:
			return
		case tx := <-w.filteredTxChannel:
			w.OutputChannel <- tx
			log.Printf("Transaction %s is sent to the OutputChannel", tx.TxID)
		}
	}
}

// getLatestBlockHeight gets the latest block height
func (w *BTCWatcher) getLatestBlockHeight() int {
	resp, err := w.client.R().Get(fmt.Sprintf("%s/blocks/tip/height", w.baseUrl))
	if err != nil {
		return 0
	}

	var height int
	if err := json.Unmarshal(resp.Body(), &height); err != nil {
		return 0
	}

	return height
}

// Close closes the btcwatcher
func (w *BTCWatcher) Close() {
	w.stopRunning <- true
	close(w.blockChannel)
	close(w.txChannel)
	close(w.filteredTxChannel)
	close(w.OutputChannel)
	close(w.stopRunning)
}
