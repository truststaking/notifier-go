package hub

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-notifier-go/common"
	"github.com/multiversx/mx-chain-notifier-go/data"
	"github.com/multiversx/mx-chain-notifier-go/dispatcher"
	"github.com/multiversx/mx-chain-notifier-go/filters"
)

var log = logger.GetOrCreate("hub")

// ArgsCommonHub defines the arguments needed for common hub creation
type ArgsCommonHub struct {
	Filter             filters.EventFilter
	SubscriptionMapper dispatcher.SubscriptionMapperHandler
}

type commonHub struct {
	filter                        filters.EventFilter
	subscriptionMapper            dispatcher.SubscriptionMapperHandler
	mutDispatchers                sync.RWMutex
	dispatchers                   map[uuid.UUID]dispatcher.EventDispatcher
	register                      chan dispatcher.EventDispatcher
	unregister                    chan dispatcher.EventDispatcher
	broadcast                     chan data.BlockEvents
	broadcastRevert               chan data.RevertBlock
	broadcastFinalized            chan data.FinalizedBlock
	broadcastTxs                  chan data.BlockTxs
	broadcastBlockEventsWithOrder chan data.BlockEventsWithOrder
	broadcastScrs                 chan data.BlockScrs
	closeChan                     chan struct{}
	cancelFunc                    func()
}

// NewCommonHub creates a new commonHub instance
func NewCommonHub(args ArgsCommonHub) (*commonHub, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &commonHub{
		mutDispatchers:                sync.RWMutex{},
		filter:                        args.Filter,
		subscriptionMapper:            args.SubscriptionMapper,
		dispatchers:                   make(map[uuid.UUID]dispatcher.EventDispatcher),
		register:                      make(chan dispatcher.EventDispatcher),
		unregister:                    make(chan dispatcher.EventDispatcher),
		broadcast:                     make(chan data.BlockEvents),
		broadcastRevert:               make(chan data.RevertBlock),
		broadcastFinalized:            make(chan data.FinalizedBlock),
		broadcastTxs:                  make(chan data.BlockTxs),
		broadcastBlockEventsWithOrder: make(chan data.BlockEventsWithOrder),
		broadcastScrs:                 make(chan data.BlockScrs),
		closeChan:                     make(chan struct{}),
	}, nil
}

func checkArgs(args ArgsCommonHub) error {
	if check.IfNil(args.Filter) {
		return ErrNilEventFilter
	}
	if check.IfNil(args.SubscriptionMapper) {
		return ErrNilSubscriptionMapper
	}

	return nil
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

		case txsEvent := <-ch.broadcastTxs:
			ch.handleTxsBroadcast(txsEvent)

		case txsEvent := <-ch.broadcastBlockEventsWithOrder:
			ch.handleBlockEventsWithOrderBroadcast(txsEvent)

		case scrsEvent := <-ch.broadcastScrs:
			ch.handleScrsBroadcast(scrsEvent)

		case dispatcherClient := <-ch.register:
			ch.registerDispatcher(dispatcherClient)

		case dispatcherClient := <-ch.unregister:
			ch.unregisterDispatcher(dispatcherClient)
		}
	}
}

// Subscribe is used by a dispatcher to send a dispatcher.SubscribeEvent
func (ch *commonHub) Subscribe(event data.SubscribeEvent) {
	ch.subscriptionMapper.MatchSubscribeEvent(event)
}

// Broadcast handles block events pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) Broadcast(events data.BlockEvents) {
	select {
	case ch.broadcast <- events:
	case <-ch.closeChan:
	}
}

// BroadcastRevert handles revert event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastRevert(event data.RevertBlock) {
	select {
	case ch.broadcastRevert <- event:
	case <-ch.closeChan:
	}
}

// BroadcastFinalized handles finalized event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastFinalized(event data.FinalizedBlock) {
	select {
	case ch.broadcastFinalized <- event:
	case <-ch.closeChan:
	}
}

// BroadcastTxs handles block txs event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastTxs(event data.BlockTxs) {
	select {
	case ch.broadcastTxs <- event:
	case <-ch.closeChan:
	}
}

// BroadcastScrs handles block scrs event pushed by producers into the broadcast channel
// Upon reading the channel, the hub notifies the registered dispatchers, if any
func (ch *commonHub) BroadcastScrs(event data.BlockScrs) {
	select {
	case ch.broadcastScrs <- event:
	case <-ch.closeChan:
	}
}

// BroadcastBlockEventsWithOrder handles full block events pushed by producers into the channel
func (ch *commonHub) BroadcastBlockEventsWithOrder(event data.BlockEventsWithOrder) {
	select {
	case ch.broadcastBlockEventsWithOrder <- event:
	case <-ch.closeChan:
	}
}

// RegisterEvent will send event to a receive-only channel used to register dispatchers
func (ch *commonHub) RegisterEvent(event dispatcher.EventDispatcher) {
	select {
	case ch.register <- event:
	case <-ch.closeChan:
	}
}

// UnregisterEvent will send event to a receive-only channel used by a dispatcher to signal it has disconnected
func (ch *commonHub) UnregisterEvent(event dispatcher.EventDispatcher) {
	select {
	case ch.unregister <- event:
	case <-ch.closeChan:
	}
}

func (ch *commonHub) handleBroadcast(blockEvents data.BlockEvents) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	for _, subscription := range subscriptions {
		if subscription.EventType != common.PushLogsAndEvents {
			continue
		}

		ch.handlePushBlockEvents(blockEvents, subscription)
	}
}

func (ch *commonHub) handlePushBlockEvents(blockEvents data.BlockEvents, subscription data.Subscription) {
	events := make([]data.Event, 0)
	for _, event := range blockEvents.Events {
		if ch.filter.MatchEvent(subscription, event) {
			events = append(events, event)
		}
	}

	ch.mutDispatchers.RLock()
	d, ok := ch.dispatchers[subscription.DispatcherID]
	if ok {
		d.PushEvents(events)
	}
	ch.mutDispatchers.RUnlock()
}

func (ch *commonHub) handleRevertBroadcast(revertBlock data.RevertBlock) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID]data.RevertBlock)

	for _, subscription := range subscriptions {
		if subscription.EventType != common.RevertBlockEvents {
			continue
		}

		dispatchersMap[subscription.DispatcherID] = revertBlock
	}

	ch.mutDispatchers.RLock()
	defer ch.mutDispatchers.RUnlock()
	for id, event := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.RevertEvent(event)
		}
	}
}

func (ch *commonHub) handleFinalizedBroadcast(finalizedBlock data.FinalizedBlock) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID]data.FinalizedBlock)

	for _, subscription := range subscriptions {
		if subscription.EventType != common.FinalizedBlockEvents {
			continue
		}

		dispatchersMap[subscription.DispatcherID] = finalizedBlock
	}

	ch.mutDispatchers.RLock()
	defer ch.mutDispatchers.RUnlock()
	for id, event := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.FinalizedEvent(event)
		}
	}
}

func (ch *commonHub) handleTxsBroadcast(blockTxs data.BlockTxs) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID]data.BlockTxs)

	for _, subscription := range subscriptions {
		if subscription.EventType != common.BlockTxs {
			continue
		}

		dispatchersMap[subscription.DispatcherID] = blockTxs
	}

	ch.mutDispatchers.RLock()
	defer ch.mutDispatchers.RUnlock()
	for id, event := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.TxsEvent(event)
		}
	}
}

func (ch *commonHub) handleBlockEventsWithOrderBroadcast(blockTxs data.BlockEventsWithOrder) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID]data.BlockEventsWithOrder)

	for _, subscription := range subscriptions {
		if subscription.EventType != common.BlockEvents {
			continue
		}

		dispatchersMap[subscription.DispatcherID] = blockTxs
	}

	ch.mutDispatchers.RLock()
	defer ch.mutDispatchers.RUnlock()
	for id, event := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.BlockEvents(event)
		}
	}
}

func (ch *commonHub) handleScrsBroadcast(blockScrs data.BlockScrs) {
	subscriptions := ch.subscriptionMapper.Subscriptions()

	dispatchersMap := make(map[uuid.UUID]data.BlockScrs)

	for _, subscription := range subscriptions {
		if subscription.EventType != common.BlockScrs {
			continue
		}

		dispatchersMap[subscription.DispatcherID] = blockScrs
	}

	ch.mutDispatchers.RLock()
	defer ch.mutDispatchers.RUnlock()
	for id, event := range dispatchersMap {
		if d, ok := ch.dispatchers[id]; ok {
			d.ScrsEvent(event)
		}
	}
}

func (ch *commonHub) registerDispatcher(d dispatcher.EventDispatcher) {
	ch.mutDispatchers.Lock()
	defer ch.mutDispatchers.Unlock()

	if _, ok := ch.dispatchers[d.GetID()]; ok {
		return
	}

	ch.dispatchers[d.GetID()] = d

	log.Info("registered new dispatcher", "dispatcherID", d.GetID())
}

func (ch *commonHub) unregisterDispatcher(d dispatcher.EventDispatcher) {
	ch.mutDispatchers.Lock()
	defer ch.mutDispatchers.Unlock()

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

	close(ch.closeChan)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ch *commonHub) IsInterfaceNil() bool {
	return ch == nil
}
