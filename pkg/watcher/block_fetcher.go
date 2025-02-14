package watcher

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/CHIHCHIEH-LAI/btcwatcher/pkg/model"
	"github.com/go-resty/resty/v2"
)

type BlockFetcher struct {
	client        *resty.Client
	baseUrl       string
	heightChannel chan *model.HeightRange
	blockChannel  chan *model.Block
	nWorkers      int
}

// NewBlockFetcher creates a new BlockFetcher instance
func NewBlockFetcher(
	baseUrl string,
	heightChannel chan *model.HeightRange,
	blockChannel chan *model.Block,
	nWorkers int,
) *BlockFetcher {

	bf := &BlockFetcher{
		client:        resty.New(),
		baseUrl:       baseUrl,
		heightChannel: heightChannel,
		blockChannel:  blockChannel,
		nWorkers:      nWorkers,
	}

	return bf
}

// Run runs the block fetcher
func (bf *BlockFetcher) Run() {
	workerPool := make(chan struct{}, bf.nWorkers) // Worker pool

	for heightRange := range bf.heightChannel {
		workerPool <- struct{}{} // Acquire a worker
		go func(height *model.HeightRange) {
			defer func() { <-workerPool }() // Release the worker
			log.Printf("Fetching blocks for height range: %d-%d", height.StartHeight, height.EndHeight)
			bf.fetchBlocks(height)
		}(heightRange)
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
