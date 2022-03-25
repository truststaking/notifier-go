package hub

import (
	"context"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/notifier-go/data"
	"github.com/ElrondNetwork/notifier-go/dispatcher"
	"github.com/ElrondNetwork/notifier-go/filters"
	"github.com/google/uuid"
)

var log = logger.GetOrCreate("hub")

type commonHub struct {
	rwMut              sync.RWMutex
	filter             filters.EventFilter
	subscriptionMapper *dispatcher.SubscriptionMapper
	dispatchers        map[uuid.UUID]dispatcher.EventDispatcher
	register           chan dispatcher.EventDispatcher
	unregister         chan dispatcher.EventDispatcher
	broadcast          chan data.BlockEvents
	broadcastRevert    chan data.RevertBlock
	broadcastFinalized chan data.FinalizedBlock
	cancelFunc         func()
}

// NewCommonHub creates a new commonHub instance
func NewCommonHub(eventFilter filters.EventFilter) *commonHub {
	return &commonHub{
		rwMut:              sync.RWMutex{},
		filter:             eventFilter,
		subscriptionMapper: dispatcher.NewSubscriptionMapper(),
		dispatchers:        make(map[uuid.UUID]dispatcher.EventDispatcher),
		register:           make(chan dispatcher.EventDispatcher),
		unregister:         make(chan dispatcher.EventDispatcher),
		broadcast:          make(chan data.BlockEvents),
		broadcastRevert:    make(chan data.RevertBlock),
		broadcastFinalized: make(chan data.FinalizedBlock),
	}
}

// Run is launched as a goroutine and listens for events on the exposed channels
func (ch *commonHub) Run() {
	var ctx context.Context
	ctx, ch.cancelFunc = context.WithCancel(context.Background())

	go ch.run(ctx)
}

func (ch *commonHub) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("commonHub is stopping...")
			return

		case events := <-ch.broadcast:
			ch.handleBroadcast(events)

		case revertEvent := <-ch.broadcastRevert:
			ch.handleRevertBroadcast(revertEvent)

		case finalizedEvent := <-ch.broadcastFinalized:
			ch.handleFinalizedBroadcast(finalizedEvent)

		case dispatcherClient := <-ch.register:
			ch.registerDispatcher(dispatcherClient)

		case dispatcherClient := <-ch.unregister:
			ch.unregisterDispatcher(dispatcherClient)
		}
	}
}

// Subscribe is used by a dispatcher to send a dispatcher.SubscribeEvent
func (ch *commonHub) Subscribe(event dispatcher.SubscribeEvent) {
	ch.subscriptionMapper.MatchSubscribeEvent(event)
}

// Broadcast handles block events pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) Broadcast(events data.BlockEvents) {
	ch.broadcast <- events
}

// BroadcastRevert handles revert event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastRevert(event data.RevertBlock) {
	ch.broadcastRevert <- event
}

// BroadcastFinalized handles finalized event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastFinalized(event data.FinalizedBlock) {
	ch.broadcastFinalized <- event
}

// RegisterEvent will send event to a receive-only channel used to register dispatchers
func (ch *commonHub) RegisterEvent(event dispatcher.EventDispatcher) {
	ch.register <- event
}

// UnregisterEvent will send event to a receive-only channel used by a dispatcher to signal it has disconnected
func (ch *commonHub) UnregisterEvent(event dispatcher.EventDispatcher) {
	ch.unregister <- event
}

func (ch *commonHub) handleBroadcast(blockEvents data.BlockEvents) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID][]data.Event)
	mapEventToDispatcher := func(id uuid.UUID, e data.Event) {
		dispatchersMap[id] = append(dispatchersMap[id], e)
	}

	for _, event := range blockEvents.Events {
		for _, subscription := range subscriptions {
			if ch.filter.MatchEvent(subscription, event) {
				mapEventToDispatcher(subscription.DispatcherID, event)
			}
		}
	}

	ch.rwMut.RLock()
	defer ch.rwMut.RUnlock()
	for id, eventValues := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.PushEvents(eventValues)
		}
	}
}

// TODO: evaluate these 2 scenarios
func (ch *commonHub) handleRevertBroadcast(revertBlock data.RevertBlock) {
}

func (ch *commonHub) handleFinalizedBroadcast(finalizedBlock data.FinalizedBlock) {
}

func (ch *commonHub) registerDispatcher(d dispatcher.EventDispatcher) {
	ch.rwMut.Lock()
	defer ch.rwMut.Unlock()

	if _, ok := ch.dispatchers[d.GetID()]; ok {
		return
	}

	ch.dispatchers[d.GetID()] = d

	log.Info("registered new dispatcher", "dispatcherID", d.GetID())
}

func (ch *commonHub) unregisterDispatcher(d dispatcher.EventDispatcher) {
	ch.rwMut.Lock()
	defer ch.rwMut.Unlock()

	if _, ok := ch.dispatchers[d.GetID()]; ok {
		delete(ch.dispatchers, d.GetID())
	}

	log.Info("unregistered dispatcher", "dispatcherID", d.GetID(), "unsubscribing", true)

	ch.subscriptionMapper.RemoveSubscriptions(d.GetID())
}

// Close will close the goroutine and channels
func (ch *commonHub) Close() error {
	if ch.cancelFunc != nil {
		ch.cancelFunc()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ch *commonHub) IsInterfaceNil() bool {
	return ch == nil
}
