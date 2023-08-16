package facade_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-notifier-go/config"
	"github.com/multiversx/mx-chain-notifier-go/data"
	"github.com/multiversx/mx-chain-notifier-go/facade"
	"github.com/multiversx/mx-chain-notifier-go/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockFacadeArgs() facade.ArgsNotifierFacade {
	return facade.ArgsNotifierFacade{
		EventsHandler:     &mocks.EventsHandlerStub{},
		APIConfig:         config.ConnectorApiConfig{},
		WSHandler:         &mocks.WSHandlerStub{},
		EventsInterceptor: &mocks.EventsInterceptorStub{},
	}
}

func TestNewNotifierFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil events handler", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()
		args.EventsHandler = nil

		f, err := facade.NewNotifierFacade(args)
		require.True(t, check.IfNil(f))
		require.Equal(t, facade.ErrNilEventsHandler, err)
	})

	t.Run("nil ws handler", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()
		args.WSHandler = nil

		f, err := facade.NewNotifierFacade(args)
		require.True(t, check.IfNil(f))
		require.Equal(t, facade.ErrNilWSHandler, err)
	})

	t.Run("nil events interceptor", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()
		args.EventsInterceptor = nil

		f, err := facade.NewNotifierFacade(args)
		require.True(t, check.IfNil(f))
		require.Equal(t, facade.ErrNilEventsInterceptor, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()
		facade, err := facade.NewNotifierFacade(args)
		require.Nil(t, err)
		require.NotNil(t, facade)
	})
}

func TestHandlePushEvents(t *testing.T) {
	t.Parallel()

	t.Run("process block events error, should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()

		expectedErr := errors.New("expected error")
		args.EventsInterceptor = &mocks.EventsInterceptorStub{
			ProcessBlockEventsCalled: func(eventsData *data.ArgsSaveBlockData) (*data.InterceptorBlockData, error) {
				return nil, expectedErr
			},
		}

		facade, err := facade.NewNotifierFacade(args)
		require.Nil(t, err)

		blockData := data.ArgsSaveBlockData{
			HeaderHash: []byte("blockHash"),
			Header:     &block.HeaderV2{},
		}
		err = facade.HandlePushEventsV2(blockData)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockFacadeArgs()

		blockHash := "blockHash1"
		txs := map[string]*data.NodeTransaction{
			"hash1": {
				TransactionHandler: &transaction.Transaction{
					Nonce: 1,
				},
				ExecutionOrder: 1,
			},
		}
		scrs := map[string]*data.NodeSmartContractResult{
			"hash2": {
				TransactionHandler: &smartContractResult.SmartContractResult{
					Nonce: 2,
				},
			},
		}
		logData := []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("logaddr1"),
					Events:  []*transaction.Event{},
				},
				TxHash: "logHash1",
			},
		}

		logEvents := []data.Event{
			{
				Address: "addr1",
			},
		}

		header := &block.HeaderV2{
			Header: &block.Header{
				ShardID: 2,
			},
		}
		blockData := data.ArgsSaveBlockData{
			HeaderHash: []byte(blockHash),
			TransactionsPool: &data.TransactionsPool{
				Txs:  txs,
				Scrs: scrs,
				Logs: logData,
			},
			Header: &block.HeaderV2{},
		}

		expTxs := map[string]*transaction.Transaction{
			"hash1": {
				Nonce: 1,
			},
		}
		expScrs := map[string]*smartContractResult.SmartContractResult{
			"hash2": {
				Nonce: 2,
			},
		}

		expTxsData := data.BlockTxs{
			Hash: blockHash,
			Txs:  expTxs,
		}
		expScrsData := data.BlockScrs{
			Hash: blockHash,
			Scrs: expScrs,
		}
		expLogEvents := data.BlockEvents{
			Hash:    blockHash,
			Events:  logEvents,
			ShardID: 2,
		}

		expTxsWithOrder := map[string]*data.NotifierTransaction{
			"hash1": {
				Transaction: &transaction.Transaction{
					Nonce: 1,
				},
				ExecutionOrder: 1,
			},
		}
		expScrsWithOrder := map[string]*data.NotifierSmartContractResult{
			"hash2": {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce: 2,
				},
			},
		}
		expTxsWithOrderData := data.BlockEventsWithOrder{
			Hash:    blockHash,
			ShardID: 2,
			Txs:     expTxsWithOrder,
			Scrs:    expScrsWithOrder,
			Events:  logEvents,
		}

		pushWasCalled := false
		txsWasCalled := false
		scrsWasCalled := false
		blockEventsWithOrderWasCalled := false
		args.EventsHandler = &mocks.EventsHandlerStub{
			HandlePushEventsCalled: func(events data.BlockEvents) error {
				pushWasCalled = true
				assert.Equal(t, expLogEvents, events)
				return nil
			},
			HandleBlockTxsCalled: func(blockTxs data.BlockTxs) {
				txsWasCalled = true
				assert.Equal(t, expTxsData, blockTxs)
			},
			HandleBlockScrsCalled: func(blockScrs data.BlockScrs) {
				scrsWasCalled = true
				assert.Equal(t, expScrsData, blockScrs)
			},
			HandleBlockEventsWithOrderCalled: func(blockTxs data.BlockEventsWithOrder) {
				blockEventsWithOrderWasCalled = true
				assert.Equal(t, expTxsWithOrderData, blockTxs)
			},
		}
		args.EventsInterceptor = &mocks.EventsInterceptorStub{
			ProcessBlockEventsCalled: func(eventsData *data.ArgsSaveBlockData) (*data.InterceptorBlockData, error) {
				return &data.InterceptorBlockData{
					Hash:          blockHash,
					Header:        header,
					Txs:           expTxs,
					Scrs:          expScrs,
					LogEvents:     logEvents,
					TxsWithOrder:  expTxsWithOrder,
					ScrsWithOrder: expScrsWithOrder,
				}, nil
			},
		}

		facade, err := facade.NewNotifierFacade(args)
		require.Nil(t, err)

		facade.HandlePushEventsV2(blockData)

		assert.True(t, pushWasCalled)
		assert.True(t, txsWasCalled)
		assert.True(t, scrsWasCalled)
		assert.True(t, blockEventsWithOrderWasCalled)
	})
}

func TestHandleRevertEvents(t *testing.T) {
	t.Parallel()

	args := createMockFacadeArgs()

	revertData := data.RevertBlock{
		Hash:  "hash1",
		Nonce: 1,
	}

	revertWasCalled := false
	args.EventsHandler = &mocks.EventsHandlerStub{
		HandleRevertEventsCalled: func(revertBlock data.RevertBlock) {
			revertWasCalled = true
			assert.Equal(t, revertData, revertBlock)
		},
	}
	facade, err := facade.NewNotifierFacade(args)
	require.Nil(t, err)

	facade.HandleRevertEvents(revertData)

	assert.True(t, revertWasCalled)
}

func TestHandleFinalizedEvents(t *testing.T) {
	t.Parallel()

	args := createMockFacadeArgs()

	finalizedData := data.FinalizedBlock{
		Hash: "hash1",
	}

	finalizedWasCalled := false
	args.EventsHandler = &mocks.EventsHandlerStub{
		HandleFinalizedEventsCalled: func(finalizedBlock data.FinalizedBlock) {
			finalizedWasCalled = true
			assert.Equal(t, finalizedData, finalizedBlock)
		},
	}
	facade, err := facade.NewNotifierFacade(args)
	require.Nil(t, err)

	facade.HandleFinalizedEvents(finalizedData)

	assert.True(t, finalizedWasCalled)
}

func TestServerHTTP(t *testing.T) {
	t.Parallel()

	args := createMockFacadeArgs()

	serveHTTPWasCalled := false
	args.WSHandler = &mocks.WSHandlerStub{
		ServeHTTPCalled: func(w http.ResponseWriter, r *http.Request) {
			serveHTTPWasCalled = true
		},
	}
	facade, err := facade.NewNotifierFacade(args)
	require.Nil(t, err)

	facade.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assert.True(t, serveHTTPWasCalled)
}

func TestGetters(t *testing.T) {
	t.Parallel()

	expuser := "user1"
	exppass := "pass1"

	args := createMockFacadeArgs()
	args.APIConfig.Username = expuser
	args.APIConfig.Password = exppass

	f, err := facade.NewNotifierFacade(args)
	require.Nil(t, err)

	user, pass := f.GetConnectorUserAndPass()
	assert.Equal(t, expuser, user)
	assert.Equal(t, exppass, pass)
}
