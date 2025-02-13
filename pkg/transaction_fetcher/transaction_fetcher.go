package transaction_fetcher

import (
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
	client *resty.Client,
	baseUrl string,
	blockChannel chan *model.Block,
	txChannel chan *model.Transaction,
	nWorkers int,
) *TransactionFetcher {
	tf := &TransactionFetcher{
		client:       client,
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
	tf.wg.Add(tf.nWorkers)
	for i := 0; i < tf.nWorkers; i++ {
		go tf.fetchTransactions()
	}
}
