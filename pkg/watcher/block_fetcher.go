package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type BlockFetcher struct {
	client        *resty.Client
	baseUrl       string
	heightChannel chan *model.HeightRange
	blockChannel  chan *model.Block
	nworkers      int
	wg            sync.WaitGroup
	stopRunning   chan struct{}
}

// NewBlockFetcher creates a new BlockFetcher instance
func NewBlockFetcher(
	baseUrl string,
	heightChannel chan *model.HeightRange,
	blockChannel chan *model.Block,
	nworkers int,
) *BlockFetcher {

	bf := &BlockFetcher{
		client:        resty.New(),
		baseUrl:       baseUrl,
		heightChannel: heightChannel,
		blockChannel:  blockChannel,
		nworkers:      nworkers,
		stopRunning:   make(chan struct{}),
	}

	return bf
}

// Run runs the block fetcher
func (bf *BlockFetcher) Run() {
	for i := 0; i < bf.nworkers; i++ {
		go bf.runWorker()
	}
}

// runWorker runs a worker that fetches blocks
func (bf *BlockFetcher) runWorker() {
	bf.wg.Add(1)
	defer bf.wg.Done()

	for {
		select {
		case heightRange := <-bf.heightChannel:
			log.Printf("Fetching blocks for height range: %d-%d", heightRange.StartHeight, heightRange.EndHeight)
			bf.fetchBlocks(heightRange)
		case <-bf.stopRunning:
			return
		}
	}
}

// fetchBlocks fetches blocks for a given height range
func (bf *BlockFetcher) fetchBlocks(height *model.HeightRange) {
	for i := height.StartHeight; i <= height.EndHeight; i++ {
		// Get the blocks starting from the given height
		resp, err := bf.client.R().Get(fmt.Sprintf("%s/blocks/%d", bf.baseUrl, i))
		if err != nil {
			continue
		}

		// Parse the response
		var subBlocks []*model.Block
		if err := json.Unmarshal(resp.Body(), &subBlocks); err != nil {
			continue
		}

		// Trim the blocks
		if height.EndHeight-i < 10 {
			subBlocks = subBlocks[:height.EndHeight-i+1]
		}

		// Send blocks to the channel
		for _, block := range subBlocks {
			bf.blockChannel <- block
		}
	}
}

// Close closes the block fetcher
func (bf *BlockFetcher) Close() {
	close(bf.stopRunning)
	bf.wg.Wait()
}
