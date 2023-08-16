package mocks

import (
	"github.com/google/uuid"
	"github.com/multiversx/mx-chain-notifier-go/data"
)

// DispatcherStub implements dispatcher EventDispatcher interface
type DispatcherStub struct {
	GetIDCalled          func() uuid.UUID
	PushEventsCalled     func(events []data.Event)
	BlockEventsCalled    func(event data.BlockEventsWithOrder)
	RevertEventCalled    func(event data.RevertBlock)
	FinalizedEventCalled func(event data.FinalizedBlock)
	TxsEventCalled       func(event data.BlockTxs)
	ScrsEventCalled      func(event data.BlockScrs)
}

// GetID -
func (d *DispatcherStub) GetID() uuid.UUID {
	if d.GetIDCalled != nil {
		return d.GetIDCalled()
	}

	return uuid.UUID{}
}

// PushEvents -
func (d *DispatcherStub) PushEvents(events []data.Event) {
	if d.PushEventsCalled != nil {
		d.PushEventsCalled(events)
	}
}

// BlockEvents -
func (d *DispatcherStub) BlockEvents(events data.BlockEventsWithOrder) {
	if d.BlockEventsCalled != nil {
		d.BlockEventsCalled(events)
	}
}

// RevertEvent -
func (d *DispatcherStub) RevertEvent(event data.RevertBlock) {
	if d.RevertEventCalled != nil {
		d.RevertEventCalled(event)
	}
}

// FinalizedEvent -
func (d *DispatcherStub) FinalizedEvent(event data.FinalizedBlock) {
	if d.FinalizedEventCalled != nil {
		d.FinalizedEventCalled(event)
	}
}

// TxsEvent -
func (d *DispatcherStub) TxsEvent(event data.BlockTxs) {
	if d.TxsEventCalled != nil {
		d.TxsEventCalled(event)
	}
}

// ScrsEvent -
func (d *DispatcherStub) ScrsEvent(event data.BlockScrs) {
	if d.ScrsEventCalled != nil {
		d.ScrsEventCalled(event)
	}
}
