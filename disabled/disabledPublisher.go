package disabled

import (
	"github.com/multiversx/mx-chain-notifier-go/data"
)

// Publisher defines a disabled publisher component
type Publisher struct{}

// Run does nothing
func (dp *Publisher) Run() {
}

// Broadcast does nothing
func (dp *Publisher) Broadcast(_ data.BlockEvents) {
}

// BroadcastRevert does nothing
func (dp *Publisher) BroadcastRevert(_ data.RevertBlock) {
}

// BroadcastFinalized does nothing
func (dp *Publisher) BroadcastFinalized(_ data.FinalizedBlock) {
}

// BroadcastTxs does nothing
func (dp *Publisher) BroadcastTxs(_ data.BlockTxs) {
}

// BroadcastScrs does nothing
func (dp *Publisher) BroadcastScrs(_ data.BlockScrs) {
}

// BroadcastBlockEventsWithOrder does nothing
func (dp *Publisher) BroadcastBlockEventsWithOrder(_ data.BlockEventsWithOrder) {
}

// Close returns nil
func (dp *Publisher) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *Publisher) IsInterfaceNil() bool {
	return dp == nil
}
