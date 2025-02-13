package transaction_fetcher

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type TransactionFetcher struct {
	client       *resty.Client
	baseUrl      string
	blockChannel chan *model.Block
	txChannel    chan *model.Transaction
	nWorkers     int
	wg           sync.WaitGroup
	stopRunning  chan bool
}

// NewTransactionFetcher creates a new TransactionFetcher instance
func NewTransactionFetcher(
	baseUrl string,
	blockChannel chan *model.Block,
	txChannel chan *model.Transaction,
	nWorkers int,
) *TransactionFetcher {
	tf := &TransactionFetcher{
		client:       resty.New(),
		baseUrl:      baseUrl,
		blockChannel: blockChannel,
		txChannel:    txChannel,
		nWorkers:     nWorkers,
		stopRunning:  make(chan bool),
	}

	return tf
}

// Run runs the transaction fetcher
func (tf *TransactionFetcher) Run() {
	for i := 0; i < tf.nWorkers; i++ {
		go tf.runWorker()
	}
}

// runWorker runs a worker that fetches transactions
func (tf *TransactionFetcher) runWorker() {
	tf.wg.Add(1)
	defer tf.wg.Done()

	for {
		select {
		case block := <-tf.blockChannel:
			tf.fetchTransactions(block)
		case <-tf.stopRunning:
			return
		}
	}
}

// fetchTransactions fetches transactions for a given block
func (tf *TransactionFetcher) fetchTransactions(block *model.Block) {
	txCount := block.TxCount
	for i := 0; i < txCount; i += 25 {
		// Fetch 25 transactions beginning at index i
		resp, err := tf.client.R().Get(fmt.Sprintf("%s/block/%s/txs/%d", tf.baseUrl, block.ID, i))
		if err != nil {
			continue
		}

		// Parse the response
		var txs []*model.Transaction
		if err := json.Unmarshal(resp.Body(), &txs); err != nil {
			continue
		}

		// Send transactions to the channel
		for _, tx := range txs {
			tf.txChannel <- tx
		}
	}
}

// Close closes the transaction fetcher
func (tf *TransactionFetcher) Close() {
	close(tf.stopRunning)
	tf.wg.Wait()
}
