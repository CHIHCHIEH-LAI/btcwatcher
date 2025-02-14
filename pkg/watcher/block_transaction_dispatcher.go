package watcher

import (
	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
)

type BlockTransactionDispatcher struct {
	blockChannel   chan *model.Block
	txRangeChannel chan *model.TransactionRange
	nWorkers       int
}

// NewBlockTransactionDispatcher creates a new BlockTransactionDispatcher instance
func NewBlockTransactionDispatcher(
	blockChannel chan *model.Block,
	txRangeChannel chan *model.TransactionRange,
	nWorkers int,
) *BlockTransactionDispatcher {
	btd := &BlockTransactionDispatcher{
		blockChannel:   blockChannel,
		txRangeChannel: txRangeChannel,
		nWorkers:       nWorkers,
	}

	return btd
}

// Run runs the block transaction dispatcher
func (btd *BlockTransactionDispatcher) Run() {
	workerPool := make(chan struct{}, btd.nWorkers) // Worker pool

	for block := range btd.blockChannel {
		workerPool <- struct{}{} // Acquire a worker
		go func(block *model.Block) {
			defer func() { <-workerPool }() // Release the worker
			btd.dispatchBlockTransactions(block)
		}(block)
	}
}

// dispatchBlockTransactions dispatches block transactions
func (btd *BlockTransactionDispatcher) dispatchBlockTransactions(block *model.Block) {
	txCount := block.TxCount
	for i := 0; i < txCount; i += 25 {
		// Send transaction range to the channel
		btd.txRangeChannel <- &model.TransactionRange{
			BlockHash: block.ID,
			StartIdx:  i,
			EndIdx:    i + 25,
		}
	}
}
