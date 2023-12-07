package process

import (
	"bytes"
	"encoding/hex"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-notifier-go/data"
)

// logEvent defines a log event associated with corresponding tx hash
type logEvent struct {
	Address      []byte
	EventHandler nodeData.EventHandler
	TxHash       string
}

// ArgsEventsInterceptor defines the arguments needed for creating an events interceptor instance
type ArgsEventsInterceptor struct {
	PubKeyConverter core.PubkeyConverter
	HexKeyConvertor core.PubkeyConverter
}

type eventsInterceptor struct {
	pubKeyConverter core.PubkeyConverter
	hexKeyConvertor core.PubkeyConverter
}

// NewEventsInterceptor creates a new eventsInterceptor instance
func NewEventsInterceptor(args ArgsEventsInterceptor) (*eventsInterceptor, error) {
	if check.IfNil(args.PubKeyConverter) {
		return nil, ErrNilPubKeyConverter
	}

	if check.IfNil(args.HexKeyConvertor) {
		return nil, ErrNilPubKeyConverter
	}

	return &eventsInterceptor{
		pubKeyConverter: args.PubKeyConverter,
		hexKeyConvertor: args.HexKeyConvertor,
	}, nil
}

// ProcessBlockEvents will process block events data
func (ei *eventsInterceptor) ProcessBlockEvents(eventsData *data.ArgsSaveBlockData) (*data.InterceptorBlockData, error) {
	if eventsData == nil {
		return nil, ErrNilBlockEvents
	}
	if eventsData.TransactionsPool == nil {
		return nil, ErrNilTransactionsPool
	}
	if eventsData.Body == nil {
		return nil, ErrNilBlockBody
	}
	if eventsData.Header == nil {
		return nil, ErrNilBlockHeader
	}
	scrs := make(map[string]*smartContractResult.SmartContractResult)
	scrHashes := make(map[string]string)
	scrsWithOrder := make(map[string]*data.NotifierSmartContractResult)
	for hash, scr := range eventsData.TransactionsPool.Scrs {
		scrs[hash] = scr.TransactionHandler
		scrHashes[hash] = hash
		scrsWithOrder[hash] = &data.NotifierSmartContractResult{
			SmartContractResult: scr.TransactionHandler,
			FeeInfo:             scr.FeeInfo,
			ExecutionOrder:      scr.ExecutionOrder,
		}
	}

	events := ei.getLogEventsFromTransactionsPool(eventsData.TransactionsPool.Logs, scrHashes)

	txs := make(map[string]*transaction.Transaction)
	txsWithOrder := make(map[string]*data.NotifierTransaction)
	for hash, tx := range eventsData.TransactionsPool.Txs {
		txs[hash] = tx.TransactionHandler
		txsWithOrder[hash] = &data.NotifierTransaction{
			Transaction:    tx.TransactionHandler,
			FeeInfo:        tx.FeeInfo,
			ExecutionOrder: tx.ExecutionOrder,
		}
	}

	return &data.InterceptorBlockData{
		Hash:          hex.EncodeToString(eventsData.HeaderHash),
		Body:          eventsData.Body,
		Header:        eventsData.Header,
		Txs:           txs,
		TxsWithOrder:  txsWithOrder,
		Scrs:          scrs,
		ScrsWithOrder: scrsWithOrder,
		LogEvents:     events,
	}, nil
}

func (ei *eventsInterceptor) getLogEventsFromTransactionsPool(logs []*data.LogData, scrs map[string]string) []data.Event {
	var logEvents []*logEvent
	for _, logData := range logs {
		if logData == nil {
			continue
		}
		if check.IfNil(logData.LogHandler) {
			continue
		}
		var tmpLogEvents []*logEvent
		skipTransfers := false
		for _, eventHandler := range logData.LogHandler.GetLogEvents() {
			eventIdentifier := string(eventHandler.GetIdentifier())
			if eventIdentifier == "signalError" || eventIdentifier == "internalVMErrors" {
				_, exists := scrs[logData.TxHash]
				
				if !exists {
					skipTransfers = true
				}
				log.Debug("eventsInterceptor: received signalError or internalVMErrors event from log event",
					"txHash", logData.TxHash,
					"isSCResult", exists,
					"skipTransfers", skipTransfers,
				)
			}
			le := &logEvent{
				EventHandler: eventHandler,
				TxHash:       logData.TxHash,
				Address:      logData.LogHandler.GetAddress(),
			}
			tmpLogEvents = append(tmpLogEvents, le)
		}
		if skipTransfers {
			var filteredItems []*logEvent
			for _, item := range tmpLogEvents {
				identifier := string(item.EventHandler.GetIdentifier())
				if identifier == "MultiESDTNFTTransfer" || identifier == "ESDTNFTTransfer" || identifier == "ESDTTransfer" {
					continue
				}
				filteredItems = append(filteredItems, item)
			}
			logEvents = append(logEvents, filteredItems...)
		} else {
			logEvents = append(logEvents, tmpLogEvents...)
		}
	}

	if len(logEvents) == 0 {
		return nil
	}
	events := make([]data.Event, 0, len(logEvents))
	for _, event := range logEvents {
		if event == nil || check.IfNil(event.EventHandler) {
			continue
		}
		hexAddress := ei.hexKeyConvertor.Encode(event.EventHandler.GetAddress())
		shardAddress := getShardOfAddress(hexAddress)
		bech32Address := ei.pubKeyConverter.Encode(event.EventHandler.GetAddress())
		bech32MainLogAddress := ei.pubKeyConverter.Encode(event.Address)
		hexMainLogAddress := ei.hexKeyConvertor.Encode(event.Address)
		shardMainLogAddress := getShardOfAddress(hexMainLogAddress)
		eventIdentifier := string(event.EventHandler.GetIdentifier())

		log.Debug("eventsInterceptor: received event from log address",
			"logAddress", bech32MainLogAddress,
			"address", bech32Address,
			"identifier", eventIdentifier,
		)

		events = append(events, data.Event{
			LogAddress:      bech32MainLogAddress,
			LogAddressShard: shardMainLogAddress,
			Address:         bech32Address,
			AddressShard:    shardAddress,
			Identifier:      eventIdentifier,
			Topics:          event.EventHandler.GetTopics(),
			Data:            event.EventHandler.GetData(),
			TxHash:          event.TxHash,
		})
	}

	return events
}

// IsInterfaceNil returns whether the interface is nil
func (ei *eventsInterceptor) IsInterfaceNil() bool {
	return ei == nil
}

const METACHAIN_SHARD_ID = 4294967295

func isAddressOfMetachain(pubKey []byte) bool {
	metachainPrefix := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	pubKeyPrefix := pubKey[:len(metachainPrefix)]
	if bytes.Equal(pubKeyPrefix, metachainPrefix) {
		return true
	}
	zeroAddress := make([]byte, 32)
	return bytes.Equal(pubKey, zeroAddress)
}

func getShardOfAddress(hexPubKey string) int {
	numShards := 3
	maskHigh := 0b11
	maskLow := 0b01

	pubKey, _ := hex.DecodeString(hexPubKey)
	lastByteOfPubKey := pubKey[31]

	if isAddressOfMetachain(pubKey) {
		return METACHAIN_SHARD_ID
	}

	shard := int(lastByteOfPubKey) & maskHigh
	if shard > numShards-1 {
		shard = int(lastByteOfPubKey) & maskLow
	}
	return shard
}
