package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/block_fetcher"
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/transaction_fetcher"
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/transaction_filter"
	"github.com/go-resty/resty/v2"
)

type BTCWatcher struct {
	client                 *resty.Client
	baseUrl                string
	lastFetchedBlockHeight int
	blockConfirmedRequired int

	heightChannel chan *model.HeightRange
	blockChannel  chan *model.Block
	txChannel     chan *model.Transaction
	OutputChannel chan *model.Transaction

	blockFetcher       *block_fetcher.BlockFetcher
	transactionFetcher *transaction_fetcher.TransactionFetcher
	transactionFilter  *transaction_filter.TransactionFilter
	stopRunning        chan bool
}

// NewBTCWatcher creates a new BTCWatcher instance
func NewBTCWatcher(network string, watchedAddresses []string, blockConfirmedRequired int) *BTCWatcher {

	w := &BTCWatcher{
		client:                 resty.New(),
		baseUrl:                getBaseUrl(network),
		lastFetchedBlockHeight: 883549, // hardcoded for now
		blockConfirmedRequired: blockConfirmedRequired,
		heightChannel:          make(chan *model.HeightRange, 10),
		blockChannel:           make(chan *model.Block, 10),
		txChannel:              make(chan *model.Transaction, 100),
		OutputChannel:          make(chan *model.Transaction, 10),
		stopRunning:            make(chan bool),
	}

	w.blockFetcher = block_fetcher.NewBlockFetcher(w.baseUrl, w.heightChannel, w.blockChannel, 10)
	w.transactionFetcher = transaction_fetcher.NewTransactionFetcher(w.baseUrl, w.blockChannel, w.txChannel, 10)
	w.transactionFilter = transaction_filter.NewTransactionFilter(watchedAddresses, w.txChannel, w.OutputChannel, 10)

	return w
}

// getBaseUrl gets the base URL for the given network
func getBaseUrl(network string) string {
	baseURL := "https://blockstream.info/api"
	if network == "testnet" {
		baseURL = "https://blockstream.info/testnet/api"
	}

	return baseURL
}

// Run starts monitoring new blocks
func (w *BTCWatcher) Run() {
	log.Println("BTCWatcher is running")

	go w.watchForNewBlock()
	go w.blockFetcher.Run()
	go w.transactionFetcher.Run()
	go w.transactionFilter.Run()
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
	// Get the latest confirmed block height
	latestConfirmedBlockHeight := w.getLatestConfirmedBlockHeight()

	for i := w.lastFetchedBlockHeight + 1; i <= latestConfirmedBlockHeight; i += 10 {
		w.heightChannel <- &model.HeightRange{
			StartHeight: i,
			EndHeight:   min(i+10, latestConfirmedBlockHeight),
		}
		// Update the last block height
		w.updatelastFetchedBlockHeight(min(i+10, latestConfirmedBlockHeight))
	}
}

// updatelastFetchedBlockHeightupdates the last block height
func (w *BTCWatcher) updatelastFetchedBlockHeight(height int) {
	w.lastFetchedBlockHeight = height
}

// getLatestConfirmedBlockHeight gets the latest confirmed block height
func (w *BTCWatcher) getLatestConfirmedBlockHeight() int {
	return w.getLatestBlockHeight() - w.blockConfirmedRequired
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
	w.transactionFetcher.Close()
	w.transactionFilter.Close()

	w.stopRunning <- true
	close(w.heightChannel)
	close(w.blockChannel)
	close(w.txChannel)
	close(w.OutputChannel)
	close(w.stopRunning)

	log.Println("BTCWatcher is closed")
}
