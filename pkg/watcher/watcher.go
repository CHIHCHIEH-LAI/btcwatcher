package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type BTCWatcher struct {
	client                 *resty.Client
	baseUrl                string
	lastFetchedBlockHeight int
	blockConfirmedRequired int

	heightChannel  chan *model.HeightRange
	blockChannel   chan *model.Block
	txRangeChannel chan *model.TransactionRange
	txChannel      chan *model.Transaction
	OutputChannel  chan *model.Transaction

	blockFetcher               *BlockFetcher
	blockTransactionDispatcher *BlockTransactionDispatcher
	transactionFetcher         *TransactionFetcher
	transactionFilter          *TransactionFilter

	wg          sync.WaitGroup
	stopRunning chan struct{}
}

// NewBTCWatcher creates a new BTCWatcher instance
func NewBTCWatcher(network string, fromHeight int, watchedAddresses []string, blockConfirmedRequired int) *BTCWatcher {

	w := &BTCWatcher{
		client:                 resty.New(),
		baseUrl:                getBaseUrl(network),
		lastFetchedBlockHeight: fromHeight,
		blockConfirmedRequired: blockConfirmedRequired,
		heightChannel:          make(chan *model.HeightRange, 10),
		blockChannel:           make(chan *model.Block, 10),
		txRangeChannel:         make(chan *model.TransactionRange, 10),
		txChannel:              make(chan *model.Transaction, 100),
		OutputChannel:          make(chan *model.Transaction, 10),
		stopRunning:            make(chan struct{}),
	}

	w.blockFetcher = NewBlockFetcher(w.baseUrl, w.heightChannel, w.blockChannel, 10)
	w.blockTransactionDispatcher = NewBlockTransactionDispatcher(w.blockChannel, w.txRangeChannel, 10)
	w.transactionFetcher = NewTransactionFetcher(w.baseUrl, w.txRangeChannel, w.txChannel, 50)
	w.transactionFilter = NewTransactionFilter(watchedAddresses, w.txChannel, w.OutputChannel, 50)

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

	w.wg.Add(5)

	go func() {
		defer w.wg.Done()
		w.watchForNewBlock()
	}()

	go func() {
		defer w.wg.Done()
		w.blockFetcher.Run()
	}()

	go func() {
		defer w.wg.Done()
		w.blockTransactionDispatcher.Run()
	}()

	go func() {
		defer w.wg.Done()
		w.transactionFetcher.Run()
	}()

	go func() {
		defer w.wg.Done()
		w.transactionFilter.Run()
	}()
}

// collectNewBlock collects new block
func (w *BTCWatcher) watchForNewBlock() {
	defer w.wg.Done()
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

// updatelastFetchedBlockHeightupdates the last block height
func (w *BTCWatcher) updatelastFetchedBlockHeight(height int) {
	w.lastFetchedBlockHeight = height
}

// Close closes the btcwatcher
func (w *BTCWatcher) Close() {
	close(w.stopRunning)

	close(w.heightChannel)
	close(w.blockChannel)
	close(w.txRangeChannel)
	close(w.txChannel)
	close(w.OutputChannel)

	w.wg.Wait()

	log.Println("BTCWatcher is closed")
}
