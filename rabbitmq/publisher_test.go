package rabbitmq_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/notifier-go/config"
	"github.com/ElrondNetwork/notifier-go/data"
	"github.com/ElrondNetwork/notifier-go/mocks"
	"github.com/ElrondNetwork/notifier-go/rabbitmq"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsRabbitMqPublisher() rabbitmq.ArgsRabbitMqPublisher {
	return rabbitmq.ArgsRabbitMqPublisher{
		Client: &mocks.RabbitClientStub{},
		Config: config.RabbitMQConfig{
			EventsExchange: config.RabbitMQExchangeConfig{
				Name: "allevents",
				Type: "fanout",
			},
			RevertEventsExchange: config.RabbitMQExchangeConfig{
				Name: "revert",
				Type: "fanout",
			},
			FinalizedEventsExchange: config.RabbitMQExchangeConfig{
				Name: "finalized",
				Type: "fanout",
			},
			BlockTxsExchange: config.RabbitMQExchangeConfig{
				Name: "blocktxs",
				Type: "fanout",
			},
			BlockScrsExchange: config.RabbitMQExchangeConfig{
				Name: "blockscrs",
				Type: "fanout",
			},
		},
	}
}

func TestRabbitMqPublisher(t *testing.T) {
	t.Parallel()

	t.Run("nil rabbitmq client", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Client = nil

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.Equal(t, rabbitmq.ErrNilRabbitMqClient, err)
	})

	t.Run("invalid events exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.EventsExchange.Name = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeName))
	})

	t.Run("invalid revert exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.RevertEventsExchange.Name = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeName))
	})

	t.Run("invalid finalized exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.FinalizedEventsExchange.Name = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeName))
	})

	t.Run("invalid txs exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.BlockTxsExchange.Name = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeName))
	})

	t.Run("invalid scrs exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.BlockScrsExchange.Name = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeName))
	})

	t.Run("invalid exchange type", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()
		args.Config.EventsExchange.Type = ""

		client, err := rabbitmq.NewRabbitMqPublisher(args)

		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeType))

		args.Config.RevertEventsExchange.Type = ""
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeType))

		args.Config.FinalizedEventsExchange.Type = ""
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeType))

		args.Config.BlockTxsExchange.Type = ""
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeType))

		args.Config.BlockScrsExchange.Type = ""
		require.True(t, check.IfNil(client))
		require.True(t, errors.Is(err, rabbitmq.ErrInvalidRabbitMqExchangeType))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsRabbitMqPublisher()

		wasCalled := false
		args.Client = &mocks.RabbitClientStub{
			ExchangeDeclareCalled: func(name, kind string) error {
				wasCalled = true
				return nil
			},
		}

		client, err := rabbitmq.NewRabbitMqPublisher(args)
		require.Nil(t, err)
		require.NotNil(t, client)
		require.True(t, wasCalled)
	})
}

func TestBroadcast(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	numCalls := uint32(0)

	client := &mocks.RabbitClientStub{
		PublishCalled: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			atomic.AddUint32(&numCalls, 1)
			wg.Done()
			return nil
		},
	}

	args := createMockArgsRabbitMqPublisher()
	args.Client = client

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()
	defer rabbitmq.Close()
	wg.Add(1)

	rabbitmq.Broadcast(data.BlockEvents{})

	wg.Wait()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestBroadcastRevert(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	numCalls := uint32(0)

	client := &mocks.RabbitClientStub{
		PublishCalled: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			atomic.AddUint32(&numCalls, 1)
			wg.Done()
			return nil
		},
	}

	args := createMockArgsRabbitMqPublisher()
	args.Client = client

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()
	defer rabbitmq.Close()
	wg.Add(1)

	rabbitmq.BroadcastRevert(data.RevertBlock{})

	wg.Wait()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestBroadcastFinalized(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	numCalls := uint32(0)

	client := &mocks.RabbitClientStub{
		PublishCalled: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			atomic.AddUint32(&numCalls, 1)
			wg.Done()
			return nil
		},
	}

	args := createMockArgsRabbitMqPublisher()
	args.Client = client

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()
	defer rabbitmq.Close()
	wg.Add(1)

	rabbitmq.BroadcastFinalized(data.FinalizedBlock{})

	wg.Wait()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestBroadcastTxs(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	numCalls := uint32(0)

	client := &mocks.RabbitClientStub{
		PublishCalled: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			atomic.AddUint32(&numCalls, 1)
			wg.Done()
			return nil
		},
	}

	args := createMockArgsRabbitMqPublisher()
	args.Client = client

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()
	defer rabbitmq.Close()
	wg.Add(1)

	rabbitmq.BroadcastTxs(data.BlockTxs{})

	wg.Wait()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestBroadcastScrs(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	numCalls := uint32(0)

	client := &mocks.RabbitClientStub{
		PublishCalled: func(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
			atomic.AddUint32(&numCalls, 1)
			wg.Done()
			return nil
		},
	}

	args := createMockArgsRabbitMqPublisher()
	args.Client = client

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()
	defer rabbitmq.Close()
	wg.Add(1)

	rabbitmq.BroadcastScrs(data.BlockScrs{})

	wg.Wait()

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestClose(t *testing.T) {
	t.Parallel()

	args := createMockArgsRabbitMqPublisher()

	rabbitmq, err := rabbitmq.NewRabbitMqPublisher(args)
	require.Nil(t, err)

	rabbitmq.Run()

	err = rabbitmq.Close()
	require.Nil(t, err)
}
