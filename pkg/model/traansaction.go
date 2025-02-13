package model

type Transaction struct {
	TxID     string `json:"txid"`
	Version  int    `json:"version"`
	Locktime int    `json:"locktime"`
	Size     int    `json:"size"`
	Weight   int    `json:"weight"`
	Fee      int    `json:"fee"`

	Vin []struct {
		TxID         string   `json:"txid"`
		Vout         int      `json:"vout"`
		IsCoinbase   bool     `json:"is_coinbase"`
		ScriptSig    string   `json:"scriptsig"`
		ScriptSigAsm string   `json:"scriptsig_asm"`
		Sequence     int64    `json:"sequence"`
		Witness      []string `json:"witness"`
		Prevout      struct {
			Address string `json:"scriptpubkey_address"`
			Value   int64  `json:"value"`
		} `json:"prevout"`
	} `json:"vin"`

	Vout []struct {
		ScriptPubKey        string `json:"scriptpubkey"`
		ScriptPubKeyAsm     string `json:"scriptpubkey_asm"`
		ScriptPubKeyType    string `json:"scriptpubkey_type"`
		ScriptPubKeyAddress string `json:"scriptpubkey_address"`
		Value               int64  `json:"value"`
	} `json:"vout"`

	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
}
