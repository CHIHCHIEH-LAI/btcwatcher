package watcher

// Block represents a Bitcoin block
type Block struct {
	ID                string  `json:"id"`                // Block hash
	Height            int     `json:"height"`            // Block height
	Version           int     `json:"version"`           // Version of the block
	Timestamp         int64   `json:"timestamp"`         // Block timestamp (Unix time)
	Bits              int     `json:"bits"`              // Compact target difficulty representation
	Nonce             int64   `json:"nonce"`             // Nonce used for mining
	Difficulty        float64 `json:"difficulty"`        // Actual mining difficulty
	MerkleRoot        string  `json:"merkle_root"`       // Root hash of transactions
	TxCount           int     `json:"tx_count"`          // Number of transactions in the block
	Size              int     `json:"size"`              // Size of the block in bytes
	Weight            int     `json:"weight"`            // Weight of the block (SegWit concept)
	PreviousBlockHash string  `json:"previousblockhash"` // Hash of the previous block
	MedianTime        int64   `json:"mediantime"`        // Median time-past for timestamp validation

	// Elements-Only Fields
	Proof        string `json:"proof,omitempty"`         // Proof for Confidential Transactions
	Challenge    string `json:"challenge,omitempty"`     // Block challenge
	ChallengeAsm string `json:"challenge_asm,omitempty"` // Challenge in assembly format
	Solution     string `json:"solution,omitempty"`      // Solution to the challenge
	SolutionAsm  string `json:"solution_asm,omitempty"`  // Solution in assembly format
}
