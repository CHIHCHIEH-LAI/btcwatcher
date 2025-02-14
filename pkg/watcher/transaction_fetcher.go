package watcher

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type TransactionFetcher struct {
	client         *resty.Client
	baseUrl        string
	txRangeChannel chan *model.TransactionRange
	txChannel      chan *model.Transaction
	nWorkers       int
	wg             sync.WaitGroup
	stopRunning    chan struct{}
}

// NewTransactionFetcher creates a new TransactionFetcher instance
func NewTransactionFetcher(
	baseUrl string,
	txRangeChannel chan *model.TransactionRange,
	txChannel chan *model.Transaction,
	nWorkers int,
) *TransactionFetcher {
	tf := &TransactionFetcher{
		client:         resty.New(),
		baseUrl:        baseUrl,
		txRangeChannel: txRangeChannel,
		txChannel:      txChannel,
		nWorkers:       nWorkers,
		stopRunning:    make(chan struct{}),
	}

	return tf
}

// Run runs the transaction fetcher
func (tf *TransactionFetcher) Run() {
	for i := 0; i < tf.nWorkers; i++ {
		tf.wg.Add(1)
		go tf.runWorker()
	}
}

// runWorker runs a worker that fetches transactions
func (tf *TransactionFetcher) runWorker() {
	defer tf.wg.Done()

	for txRange := range tf.txRangeChannel {
		select {
		case <-tf.stopRunning:
			return
		default:
			tf.fetchTransactions(txRange)
		}
	}
}

// fetchTransactions fetches transactions for a given block
func (tf *TransactionFetcher) fetchTransactions(txRange *model.TransactionRange) {
	blockHash := txRange.BlockHash
	startIdx := txRange.StartIdx
	endIdx := txRange.EndIdx
	resp, err := tf.client.R().Get(fmt.Sprintf("%s/block/%s/txs/%d", tf.baseUrl, blockHash, startIdx))
	if err != nil {
		return
	}

	// Parse the response
	var txs []*model.Transaction
	if err := json.Unmarshal(resp.Body(), &txs); err != nil {
		return
	}

	// Trim the transactions
	if endIdx < len(txs) {
		txs = txs[:endIdx]
	}

	// Send transactions to the channel
	for _, tx := range txs {
		tf.txChannel <- tx
	}
}

// Close closes the transaction fetcher
func (tf *TransactionFetcher) Close() {
	close(tf.stopRunning)
	tf.wg.Wait()
}
