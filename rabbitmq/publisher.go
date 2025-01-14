package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"

	"bytes"
	"encoding/hex"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-notifier-go/config"
	"github.com/multiversx/mx-chain-notifier-go/data"
	"github.com/streadway/amqp"
)

const (
	emptyStr = ""
)

var log = logger.GetOrCreate("rabbitmq")

// ArgsRabbitMqPublisher defines the arguments needed for rabbitmq publisher creation
type ArgsRabbitMqPublisher struct {
	Client RabbitMqClient
	Config config.RabbitMQConfig
}

type rabbitMqPublisher struct {
	client RabbitMqClient
	cfg    config.RabbitMQConfig

	broadcast                     chan data.BlockEvents
	broadcastRevert               chan data.RevertBlock
	broadcastFinalized            chan data.FinalizedBlock
	broadcastTxs                  chan data.BlockTxs
	broadcastBlockEventsWithOrder chan data.BlockEventsWithOrder
	broadcastScrs                 chan data.BlockScrs

	azure      *azservicebus.Client
	topic      string
	cancelFunc func()
	closeChan  chan struct{}
}

// NewRabbitMqPublisher creates a new rabbitMQ publisher instance
func NewRabbitMqPublisher(args ArgsRabbitMqPublisher) (*rabbitMqPublisher, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	client, err := azservicebus.NewClientFromConnectionString(args.Config.AzureCredentials, nil)

	if err != nil {
		return nil, err
	}

	rp := &rabbitMqPublisher{
		broadcast:                     make(chan data.BlockEvents),
		broadcastRevert:               make(chan data.RevertBlock),
		broadcastFinalized:            make(chan data.FinalizedBlock),
		broadcastTxs:                  make(chan data.BlockTxs),
		broadcastScrs:                 make(chan data.BlockScrs),
		broadcastBlockEventsWithOrder: make(chan data.BlockEventsWithOrder),
		cfg:                           args.Config,
		client:                        args.Client,
		closeChan:                     make(chan struct{}),
		azure:                         client,
		topic:                         args.Config.Topic,
	}

	// err = rp.createExchanges()
	// if err != nil {
	// 	return nil, err
	// }

	return rp, nil
}

func checkArgs(args ArgsRabbitMqPublisher) error {
	if check.IfNil(args.Client) {
		return ErrNilRabbitMqClient
	}

	if args.Config.EventsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.EventsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}
	if args.Config.RevertEventsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.RevertEventsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}
	if args.Config.FinalizedEventsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.FinalizedEventsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}
	if args.Config.BlockTxsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.BlockTxsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}
	if args.Config.BlockScrsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.BlockScrsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}
	if args.Config.BlockEventsExchange.Name == "" {
		return ErrInvalidRabbitMqExchangeName
	}
	if args.Config.BlockEventsExchange.Type == "" {
		return ErrInvalidRabbitMqExchangeType
	}

	return nil
}

// checkAndCreateExchanges creates exchanges if they are not existing already
// func (rp *rabbitMqPublisher) createExchanges() error {
// 	err := rp.createExchange(rp.cfg.EventsExchange)
// 	if err != nil {
// 		return err
// 	}
// 	err = rp.createExchange(rp.cfg.RevertEventsExchange)
// 	if err != nil {
// 		return err
// 	}
// 	err = rp.createExchange(rp.cfg.FinalizedEventsExchange)
// 	if err != nil {
// 		return err
// 	}
// 	err = rp.createExchange(rp.cfg.BlockTxsExchange)
// 	if err != nil {
// 		return err
// 	}
// 	err = rp.createExchange(rp.cfg.BlockScrsExchange)
// 	if err != nil {
// 		return err
// 	}
// 	err = rp.createExchange(rp.cfg.BlockEventsExchange)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (rp *rabbitMqPublisher) createExchange(conf config.RabbitMQExchangeConfig) error {
// 	err := rp.client.ExchangeDeclare(conf.Name, conf.Type)
// 	if err != nil {
// 		return err
// 	}

// 	log.Info("checked and declared rabbitMQ exchange", "name", conf.Name, "type", conf.Type)

// 	return nil
// }

// Run is launched as a goroutine and listens for events on the exposed channels
func (rp *rabbitMqPublisher) Run() {
	var ctx context.Context
	ctx, rp.cancelFunc = context.WithCancel(context.Background())

	go rp.run(ctx)
}

func (rp *rabbitMqPublisher) run(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			log.Debug("RabbitMQ publisher is stopping...")
			rp.client.Close()
		case events := <-rp.broadcast:
			rp.publishToExchanges(events)
		case revertBlock := <-rp.broadcastRevert:
			rp.publishRevertToExchange(revertBlock)
		case finalizedBlock := <-rp.broadcastFinalized:
			rp.publishFinalizedToExchange(finalizedBlock)
		case blockTxs := <-rp.broadcastTxs:
			rp.publishTxsToExchange(blockTxs)
		case blockScrs := <-rp.broadcastScrs:
			rp.publishScrsToExchange(blockScrs)
		case blockEvents := <-rp.broadcastBlockEventsWithOrder:
			rp.publishBlockEventsWithOrderToExchange(blockEvents)
		case err := <-rp.client.ConnErrChan():
			if err != nil {
				log.Error("rabbitMQ connection failure", "err", err.Error())
				rp.client.Reconnect()
			}
		case err := <-rp.client.CloseErrChan():
			if err != nil {
				log.Error("rabbitMQ channel failure", "err", err.Error())
				rp.client.ReopenChannel()
			}
		}
	}
}

// Broadcast will handle the block events pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) Broadcast(events data.BlockEvents) {
	select {
	case rp.broadcast <- events:
	case <-rp.closeChan:
	}
}

// BroadcastRevert will handle the revert event pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) BroadcastRevert(events data.RevertBlock) {
	select {
	case rp.broadcastRevert <- events:
	case <-rp.closeChan:
	}
}

// BroadcastFinalized will handle the finalized event pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) BroadcastFinalized(events data.FinalizedBlock) {
	select {
	case rp.broadcastFinalized <- events:
	case <-rp.closeChan:
	}
}

// BroadcastTxs will handle the txs event pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) BroadcastTxs(events data.BlockTxs) {
	select {
	case rp.broadcastTxs <- events:
	case <-rp.closeChan:
	}
}

// BroadcastScrs will handle the scrs event pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) BroadcastScrs(events data.BlockScrs) {
	select {
	case rp.broadcastScrs <- events:
	case <-rp.closeChan:
	}
}

// BroadcastBlockEventsWithOrder will handle the full block events pushed by producers and sends them to rabbitMQ channel
func (rp *rabbitMqPublisher) BroadcastBlockEventsWithOrder(events data.BlockEventsWithOrder) {
	select {
	case rp.broadcastBlockEventsWithOrder <- events:
	case <-rp.closeChan:
	}
}

func (rp *rabbitMqPublisher) publishToExchanges(events data.BlockEvents) {
	eventsBytes, err := json.Marshal(events)
	if err != nil {
		log.Error("could not marshal events", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.EventsExchange.Name, eventsBytes)
	if err != nil {
		log.Error("failed to publish events to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishRevertToExchange(revertBlock data.RevertBlock) {
	revertBlockBytes, err := json.Marshal(revertBlock)
	if err != nil {
		log.Error("could not marshal revert event", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.RevertEventsExchange.Name, revertBlockBytes)
	if err != nil {
		log.Error("failed to publish revert event to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishFinalizedToExchange(finalizedBlock data.FinalizedBlock) {
	finalizedBlockBytes, err := json.Marshal(finalizedBlock)
	if err != nil {
		log.Error("could not marshal finalized event", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.FinalizedEventsExchange.Name, finalizedBlockBytes)
	if err != nil {
		log.Error("failed to publish finalized event to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishTxsToExchange(blockTxs data.BlockTxs) {
	txsBlockBytes, err := json.Marshal(blockTxs)
	if err != nil {
		log.Error("could not marshal block txs event", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.BlockTxsExchange.Name, txsBlockBytes)
	if err != nil {
		log.Error("failed to publish block txs event to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishScrsToExchange(blockScrs data.BlockScrs) {
	scrsBlockBytes, err := json.Marshal(blockScrs)
	if err != nil {
		log.Error("could not marshal block scrs event", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.BlockScrsExchange.Name, scrsBlockBytes)
	if err != nil {
		log.Error("failed to publish block scrs event to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishBlockEventsWithOrderToExchange(blockTxs data.BlockEventsWithOrder) {
	txsBlockBytes, err := json.Marshal(blockTxs)
	if err != nil {
		log.Error("could not marshal block txs event", "err", err.Error())
		return
	}

	err = rp.publishFanout(rp.cfg.BlockEventsExchange.Name, txsBlockBytes)
	if err != nil {
		log.Error("failed to publish full block events to rabbitMQ", "err", err.Error())
	}
}

func (rp *rabbitMqPublisher) publishFanout(exchangeName string, payload []byte) error {

	if exchangeName == rp.cfg.BlockEventsExchange.Name {
		sender, err := rp.azure.NewSender(rp.cfg.Topic, nil)
		if err != nil {
			log.Error("could not send the payload to azure service bus", "err", err.Error())
			return err
		}
		var events data.BlockEventsWithOrder
		err = json.Unmarshal(payload, &events)
		if err != nil {
			log.Error("Error unmarshalling JSON data for service bus:", err)
		}
		currentMessageBatch, err := sender.NewMessageBatch(context.Background(), nil)

		if err != nil {
			log.Error("error creating message batch for service bus:", err)
		}
		for i := 0; i < len(events.Events); i++ {
			identifier := events.Events[i].Identifier
			sessionId := events.Events[i].Address
			isNFT := "true"
			if identifier == "completedTxEvent" || identifier == "signalError" || identifier == "internalVMErrors" || identifier == "writeLog" {
				continue
			}

			if identifier == "MultiESDTNFTTransfer" || identifier == "ESDTNFTTransfer" || identifier == "ESDTTransfer" {
				pubKeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(32)
				if err != nil {
					return err
				}
				reciver := pubKeyConverter.Encode(events.Events[i].Topics[3])
				receiverShard := getShardOfAddress(reciver)

				if receiverShard != events.Events[i].LogAddressShard {
					continue
				}
			}

			if identifier == "ESDTNFTCreate" || identifier == "ESDTNFTBurn" || identifier == "ESDTNFTUpdateAttributes" || identifier == "ESDTNFTAddURI" || identifier == "ESDTNFTAddQuantity" || identifier == "MultiESDTNFTTransfer" || identifier == "ESDTNFTTransfer" || identifier == "ESDTTransfer" {
				hexStr := hex.EncodeToString(events.Events[i].Topics[1])
				if hexStr == ""  {
					isNFT = "false"
				}
				sessionId = string(events.Events[i].Topics[0])
			}

			event, err := json.Marshal(events.Events[i])
			if err != nil {
				log.Error("Error marshalling JSON data for service bus:", err)
			}

			msg := &azservicebus.Message{
				Body:                  event,
				SessionID:             &sessionId,
				ApplicationProperties: make(map[string]interface{})}
			msg.ApplicationProperties["Address"] = events.Events[i].Address

			if identifier == "MultiESDTNFTTransfer" {
				msg.ApplicationProperties["isNFT"] = isNFT
			}

			msg.ApplicationProperties["Identifier"] = events.Events[i].Identifier
			err = currentMessageBatch.AddMessage(msg, nil)

			if errors.Is(err, azservicebus.ErrMessageTooLarge) {
				if currentMessageBatch.NumMessages() == 0 {
					log.Error("Single message is too large to be sent in a batch.")
					return err
				}

				log.Info("Message batch is full. Sending it and creating a new one.")

				// send what we have since the batch is full
				err := sender.SendMessageBatch(context.Background(), currentMessageBatch, nil)

				if err != nil {
					log.Error("Error sending the batch of messages", err)
					return err
				}

				// Create a new batch and retry adding this message to our batch.
				newBatch, err := sender.NewMessageBatch(context.Background(), nil)

				if err != nil {
					log.Error("Error creating a new batch of messages", err)
				}

				currentMessageBatch = newBatch

				// rewind the counter and attempt to add the message again (this batch
				// was full so it didn't go out with the previous SendMessageBatch call).
				i--
			} else if err != nil {
				log.Error("Error adding message to batch", currentMessageBatch.NumMessages(), err)
			}
		}
		// check if any messages are remaining to be sent.
		if currentMessageBatch.NumMessages() > 0 {
			err := sender.SendMessageBatch(context.Background(), currentMessageBatch, nil)

			if err != nil {
				log.Error("Error send remaining messages in batch", err)
			}
		}
		sender.Close(context.Background())
	}

	return rp.client.Publish(
		exchangeName,
		emptyStr,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			Body: payload,
		},
	)
}

// Close will close the channels
func (rp *rabbitMqPublisher) Close() error {
	if rp.cancelFunc != nil {
		rp.cancelFunc()
	}

	close(rp.closeChan)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rp *rabbitMqPublisher) IsInterfaceNil() bool {
	return rp == nil
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
