package data

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/core"
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// WSEvent defines a websocket event
type WSEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// TxLog holds log data
type TxLog struct {
	Address string  `json:"scAddress"`
	Events  []Event `json:"events"`
}

// LogEvent defines a log event associated with corresponding tx hash
type LogEvent struct {
	EventHandler nodeData.EventHandler
	TxHash       string
}

// Event holds event data
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
	TxHash     string   `json:"txHash"`
}

// BlockEvents holds events data for a block
type BlockEvents struct {
	Hash   string  `json:"hash"`
	Events []Event `json:"events"`
}

// RevertBlock holds revert event data
type RevertBlock struct {
	Hash  string `json:"hash"`
	Nonce uint64 `json:"nonce"`
	Round uint64 `json:"round"`
	Epoch uint32 `json:"epoch"`
}

// FinalizedBlock holds finalized block data
type FinalizedBlock struct {
	Hash string `json:"hash"`
}

// BlockTxs holds the block transactions
// TODO: set transaction with order here also
type BlockTxs struct {
	Hash string                              `json:"hash"`
	Txs  map[string]*transaction.Transaction `json:"txs"`
}

// BlockScrs holds the block smart contract results
// TODO: set scr with order here also
type BlockScrs struct {
	Hash string                                              `json:"hash"`
	Scrs map[string]*smartContractResult.SmartContractResult `json:"scrs"`
}

// SaveBlockData holds the filtered block data that will be received on push events
type SaveBlockData struct {
	Hash      string                                              `json:"hash"`
	Txs       map[string]*transaction.Transaction                 `json:"txs"`
	Scrs      map[string]*smartContractResult.SmartContractResult `json:"scrs"`
	LogEvents []Event                                             `json:"events"`
}

// ArgsSaveBlockData holds the block data that will be received on push events
type ArgsSaveBlockData struct {
	HeaderHash             []byte
	Body                   nodeData.BodyHandler
	Header                 nodeData.HeaderHandler
	SignersIndexes         []uint64
	NotarizedHeadersHashes []string
	HeaderGasConsumption   outport.HeaderGasConsumption
	TransactionsPool       *TransactionsPool
	AlteredAccounts        map[string]*outport.AlteredAccount
	NumberOfShards         uint32
	IsImportDB             bool
}

// ArgsSaveBlock holds block data with header type
type ArgsSaveBlock struct {
	HeaderType core.HeaderType
	ArgsSaveBlockData
}

// LogData holds the data needed for indexing logs and events
type LogData struct {
	LogHandler *transaction.Log
	TxHash     string
}

// TransactionsPool holds all types of transaction
type TransactionsPool struct {
	Txs      map[string]TransactionWithOrder
	Scrs     map[string]SmartContractResultWithOrder
	Rewards  map[string]RewardTxWithOrder
	Invalid  map[string]TransactionWithOrder
	Receipts map[string]ReceiptWithOrder
	Logs     []*LogData
}

// TransactionWithOrder defines a wrapper over transaction
type TransactionWithOrder struct {
	TransactionHandler *transaction.Transaction
	ExecutionOrder     int
}

// SmartContractResultWithOrder defines a wrapper over scr
type SmartContractResultWithOrder struct {
	TransactionHandler *smartContractResult.SmartContractResult
	ExecutionOrder     int
}

// RewardTxWithOrder defines a wrapper over rewardTx
type RewardTxWithOrder struct {
	TransactionHandler *rewardTx.RewardTx
	ExecutionOrder     int
}

// ReceiptWithOrder defines a wrapper over receipt
type ReceiptWithOrder struct {
	TransactionHandler *receipt.Receipt
	ExecutionOrder     int
}
