package watcher

import (
	"encoding/json"
	"fmt"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type TransactionFetcher struct {
	client         *resty.Client
	baseUrl        string
	txRangeChannel chan *model.TransactionRange
	txChannel      chan *model.Transaction
	nWorkers       int
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
	}

	return tf
}

// Run runs the transaction fetcher
func (tf *TransactionFetcher) Run() {
	workerPool := make(chan struct{}, tf.nWorkers) // Worker pool

	for txRange := range tf.txRangeChannel {
		workerPool <- struct{}{} // Acquire a worker
		go func(txRange *model.TransactionRange) {
			defer func() { <-workerPool }() // Release the worker
			// log.Printf("Fetching transactions for block: %s, range: %d-%d", txRange.BlockHash, txRange.StartIdx, txRange.EndIdx)
			tf.fetchTransactions(txRange)
		}(txRange)
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
