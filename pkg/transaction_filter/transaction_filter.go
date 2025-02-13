package transaction_filter

import (
	"sync"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
)

type TransactionFilter struct {
	watchedAddresses  map[string]bool
	txChannel         chan *model.Transaction
	filteredTxChannel chan *model.Transaction
	nWorkers          int
	wg                sync.WaitGroup
	stopRunning       chan struct{}
}

// NewTransactionFilter creates a new TransactionFilter instance
func NewTransactionFilter(
	watchedAddresses []string,
	txChannel chan *model.Transaction,
	filteredTxChannel chan *model.Transaction,
	nWorkers int,
) *TransactionFilter {
	tf := &TransactionFilter{
		txChannel:         txChannel,
		filteredTxChannel: filteredTxChannel,
		nWorkers:          nWorkers,
		stopRunning:       make(chan struct{}),
	}

	tf.setWatchedAddresses(watchedAddresses)

	return tf
}

// setWatchedAddresses sets the addresses to watch
func (tf *TransactionFilter) setWatchedAddresses(addresses []string) {
	tf.watchedAddresses = make(map[string]bool)
	for _, addr := range addresses {
		tf.watchedAddresses[addr] = true
	}
}

// Run runs the transaction filter
func (tf *TransactionFilter) Run() {
	for i := 0; i < tf.nWorkers; i++ {
		go tf.runWorker()
	}
}

func (tf *TransactionFilter) runWorker() {
	tf.wg.Add(1)
	defer tf.wg.Done()

	for {
		select {
		case tx := <-tf.txChannel:
			tf.filterTransaction(tx)
		case <-tf.stopRunning:
			return
		}
	}
}

// filterTransaction filters transaction
func (tf *TransactionFilter) filterTransaction(tx *model.Transaction) {
	if tf.isTransactionWatched(tx) {
		tf.filteredTxChannel <- tx
	}
}

// isTransactionWatched checks if a transaction is watched
func (tf *TransactionFilter) isTransactionWatched(tx *model.Transaction) bool {
	for _, vout := range tx.Vout {
		if tf.watchedAddresses[vout.ScriptPubKeyAddress] {
			return true
		}
	}

	return false
}

// Close closes the transaction filter
func (tf *TransactionFilter) Close() {
	close(tf.stopRunning)
	tf.wg.Wait()
}
