package facade

import (
	"github.com/multiversx/mx-chain-notifier-go/data"
	"github.com/multiversx/mx-chain-notifier-go/dispatcher"
)

// EventsHandler defines the behavior of an events handler component.
// This will handle push events from observer node.
type EventsHandler interface {
	HandlePushEvents(events data.BlockEvents) error
	HandleRevertEvents(revertBlock data.RevertBlock)
	HandleFinalizedEvents(finalizedBlock data.FinalizedBlock)
	HandleBlockTxs(blockTxs data.BlockTxs)
	HandleBlockScrs(blockScrs data.BlockScrs)
	HandleBlockEventsWithOrder(blockTxs data.BlockEventsWithOrder)
	IsInterfaceNil() bool
}

// HubHandler defines the behaviour of a hub component which should be able to register
// and unregister dispatching events
type HubHandler interface {
	Publisher
	Run()
	RegisterEvent(event dispatcher.EventDispatcher)
	UnregisterEvent(event dispatcher.EventDispatcher)
	Subscribe(event data.SubscribeEvent)
	Close() error
	IsInterfaceNil() bool
}

// Publisher defines the behaviour of a publisher component which should be
// able to publish received events and broadcast them to channels
type Publisher interface {
	Broadcast(events data.BlockEvents)
	BroadcastRevert(event data.RevertBlock)
	BroadcastFinalized(event data.FinalizedBlock)
	IsInterfaceNil() bool
}

// EventsInterceptor defines the behaviour of an events interceptor component
type EventsInterceptor interface {
	ProcessBlockEvents(eventsData *data.ArgsSaveBlockData) (*data.InterceptorBlockData, error)
	IsInterfaceNil() bool
}
