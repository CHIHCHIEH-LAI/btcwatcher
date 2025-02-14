package watcher

import (
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
)

type TransactionFilter struct {
	watchedAddresses  map[string]bool
	txChannel         chan *model.Transaction
	filteredTxChannel chan *model.Transaction
	nWorkers          int
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
	workerPool := make(chan struct{}, tf.nWorkers)

	for tx := range tf.txChannel {
		workerPool <- struct{}{} // Acquire a worker
		go func(tx *model.Transaction) {
			defer func() { <-workerPool }() // Release the worker
			// log.Printf("Filtering transaction: %s", tx.TxID)
			tf.filterTransaction(tx)
		}(tx)
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
