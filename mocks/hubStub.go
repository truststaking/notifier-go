package mocks

import (
	"github.com/multiversx/mx-chain-notifier-go/data"
	"github.com/multiversx/mx-chain-notifier-go/dispatcher"
)

// HubStub implements Hub interface
type HubStub struct {
	RunCalled                func()
	BroadcastCalled          func(events data.BlockEvents)
	BroadcastRevertCalled    func(event data.RevertBlock)
	BroadcastFinalizedCalled func(event data.FinalizedBlock)
	BroadcastTxsCalled       func(event data.BlockTxs)
	BroadcastScrsCalled      func(event data.BlockScrs)
	RegisterEventCalled      func(event dispatcher.EventDispatcher)
	UnregisterEventCalled    func(event dispatcher.EventDispatcher)
	SubscribeCalled          func(event data.SubscribeEvent)
	CloseCalled              func() error
}

// Run -
func (h *HubStub) Run() {
	if h.RunCalled != nil {
		h.RunCalled()
	}
}

// Broadcast -
func (h *HubStub) Broadcast(events data.BlockEvents) {
	if h.BroadcastCalled != nil {
		h.BroadcastCalled(events)
	}
}

// BroadcastRevert -
func (h *HubStub) BroadcastRevert(event data.RevertBlock) {
	if h.BroadcastRevertCalled != nil {
		h.BroadcastRevertCalled(event)
	}
}

// BroadcastFinalized -
func (h *HubStub) BroadcastFinalized(event data.FinalizedBlock) {
	if h.BroadcastFinalizedCalled != nil {
		h.BroadcastFinalizedCalled(event)
	}
}

// BroadcastTxs -
func (h *HubStub) BroadcastTxs(event data.BlockTxs) {
	if h.BroadcastTxsCalled != nil {
		h.BroadcastTxsCalled(event)
	}
}

// BroadcastScrs -
func (h *HubStub) BroadcastScrs(event data.BlockScrs) {
	if h.BroadcastScrsCalled != nil {
		h.BroadcastScrsCalled(event)
	}
}

// RegisterEvent -
func (h *HubStub) RegisterEvent(event dispatcher.EventDispatcher) {
	if h.RegisterEventCalled != nil {
		h.RegisterEventCalled(event)
	}
}

// UnregisterEvent -
func (h *HubStub) UnregisterEvent(event dispatcher.EventDispatcher) {
	if h.UnregisterEventCalled != nil {
		h.UnregisterEventCalled(event)
	}
}

// Subscribe -
func (h *HubStub) Subscribe(event data.SubscribeEvent) {
	if h.SubscribeCalled != nil {
		h.SubscribeCalled(event)
	}
}

// Close -
func (h *HubStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (h *HubStub) IsInterfaceNil() bool {
	return h == nil
}
