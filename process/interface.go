package process

import (
	"context"

	"github.com/multiversx/mx-chain-notifier-go/data"
)

// LockService defines the behaviour of a lock service component.
// It makes sure that a duplicated entry is not processed multiple times.
type LockService interface {
	IsEventProcessed(ctx context.Context, blockHash string) (bool, error)
	HasConnection(ctx context.Context) bool
	IsInterfaceNil() bool
}

// Publisher defines the behaviour of a publisher component which should be
// able to publish received events and broadcast them to channels
type Publisher interface {
	Broadcast(events data.BlockEvents)
	BroadcastRevert(event data.RevertBlock)
	BroadcastFinalized(event data.FinalizedBlock)
	BroadcastTxs(event data.BlockTxs)
	BroadcastScrs(event data.BlockScrs)
	IsInterfaceNil() bool
}

// EventsHandler defines the behaviour of an events handler component
type EventsHandler interface {
	HandlePushEvents(events data.BlockEvents) error
	HandleRevertEvents(revertBlock data.RevertBlock)
	HandleFinalizedEvents(finalizedBlock data.FinalizedBlock)
	HandleBlockTxs(blockTxs data.BlockTxs)
	HandleBlockScrs(blockScrs data.BlockScrs)
	IsInterfaceNil() bool
}

// EventsInterceptor defines the behaviour of an events interceptor component
type EventsInterceptor interface {
	ProcessBlockEvents(eventsData *data.ArgsSaveBlockData) (*data.SaveBlockData, error)
	IsInterfaceNil() bool
}
