package watcher

import (
	"sync"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
)

type BlockTransactionDispatcher struct {
	blockChannel   chan *model.Block
	txRangeChannel chan *model.TransactionRange
	nWorkers       int
	wg             sync.WaitGroup
	stopRunning    chan struct{}
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
		stopRunning:    make(chan struct{}),
	}

	return btd
}

// Run runs the block transaction dispatcher
func (btd *BlockTransactionDispatcher) Run() {
	for i := 0; i < btd.nWorkers; i++ {
		btd.wg.Add(1)
		go btd.runWorker()
	}
}

// runWorker runs a worker that dispatches block transactions
func (btd *BlockTransactionDispatcher) runWorker() {
	defer btd.wg.Done()

	for block := range btd.blockChannel {
		select {
		case <-btd.stopRunning:
			return
		default:
			btd.dispatchBlockTransactions(block)
		}
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

// Close closes the block transaction dispatcher
func (btd *BlockTransactionDispatcher) Close() {
	close(btd.stopRunning)
	btd.wg.Wait()
}
