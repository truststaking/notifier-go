package handlers

import (
	"github.com/ElrondNetwork/notifier-go/config"
	"github.com/ElrondNetwork/notifier-go/data"
	"github.com/ElrondNetwork/notifier-go/dispatcher"
	"github.com/ElrondNetwork/notifier-go/pubsub"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	baseEventsEndpoint = "/events"
	pushEventsEndpoint = "/push"

	setnxRetryMs = 500
)

type eventsHandler struct {
	notifierHub dispatcher.Hub
	config      config.ConnectorApiConfig
	redlock     *pubsub.RedlockWrapper
}

// NewEventsHandler registers handlers for the /events group
func NewEventsHandler(
	notifierHub dispatcher.Hub,
	groupHandler *GroupHandler,
	config config.ConnectorApiConfig,
	redlock *pubsub.RedlockWrapper,
) error {
	h := &eventsHandler{
		notifierHub: notifierHub,
		config:      config,
		redlock:     redlock,
	}

	endpoints := []EndpointHandler{
		{Method: http.MethodPost, Path: pushEventsEndpoint, HandlerFunc: h.pushEvents},
	}

	endpointGroupHandler := EndpointGroupHandler{
		Root:             baseEventsEndpoint,
		Middlewares:      h.createMiddlewares(),
		EndpointHandlers: endpoints,
	}

	groupHandler.AddEndpointGroupHandler(endpointGroupHandler)

	return nil
}

func (h *eventsHandler) pushEvents(c *gin.Context) {
	var blockEvents data.BlockEvents

	err := c.Bind(&blockEvents)
	if err != nil {
		JsonResponse(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	shouldProcessEvents := true
	if h.config.CheckDuplicates {
		shouldProcessEvents = h.tryCheckProcessedOrRetry(blockEvents.Hash)
	}

	if blockEvents.Events != nil && shouldProcessEvents {
		log.Info("received events for block",
			"block hash", string(blockEvents.Hash),
			"shouldProcess", shouldProcessEvents,
		)
		h.notifierHub.BroadcastChan() <- blockEvents.Events
	}

	JsonResponse(c, http.StatusOK, nil, "")
}

func (h *eventsHandler) createMiddlewares() []gin.HandlerFunc {
	var middleware []gin.HandlerFunc

	if h.config.Username != "" && h.config.Password != "" {
		basicAuth := gin.BasicAuth(gin.Accounts{
			h.config.Username: h.config.Password,
		})
		middleware = append(middleware, basicAuth)
	}

	return middleware
}

func (h *eventsHandler) tryCheckProcessedOrRetry(blockHash []byte) bool {
	var err error
	var processed bool

	for {
		processed, err = h.redlock.IsBlockProcessed(blockHash)

		if err != nil {
			if netErr, ok := err.(*net.OpError); ok {
				if strings.Contains(netErr.Err.Error(), "connection refused") {
					time.Sleep(time.Millisecond * setnxRetryMs)
					continue
				}
			}
		}

		break
	}

	return err == nil && processed
}