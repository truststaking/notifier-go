package groups

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-notifier-go/api/errors"
	"github.com/multiversx/mx-chain-notifier-go/api/shared"
	"github.com/multiversx/mx-chain-notifier-go/data"
)

const (
	pushEventsEndpoint      = "/push"
	revertEventsEndpoint    = "/revert"
	finalizedEventsEndpoint = "/finalized"
)

type eventsGroup struct {
	*baseGroup
	facade EventsFacadeHandler
}

// NewEventsGroup registers handlers for the /events group
func NewEventsGroup(facade EventsFacadeHandler) (*eventsGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for events group", errors.ErrNilFacadeHandler)
	}

	h := &eventsGroup{
		facade:    facade,
		baseGroup: newBaseGroup(),
	}

	h.createMiddlewares()

	endpoints := []*shared.EndpointHandlerData{
		{
			Method:  http.MethodPost,
			Path:    pushEventsEndpoint,
			Handler: h.pushEvents,
		},
		{
			Method:  http.MethodPost,
			Path:    revertEventsEndpoint,
			Handler: h.revertEvents,
		},
		{
			Method:  http.MethodPost,
			Path:    finalizedEventsEndpoint,
			Handler: h.finalizedEvents,
		},
	}

	h.endpoints = endpoints

	return h, nil
}

func (h *eventsGroup) pushEvents(c *gin.Context) {
	pushEventsRawData, err := c.GetRawData()
	if err != nil {
		shared.JSONResponse(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	// blockEvents, err := UnmarshallBlockDataV1(pushEventsRawData)
	// if err == nil {
	// 	err = h.facade.HandlePushEventsV1(*blockEvents)
	// 	if err == nil {
	// 		shared.JSONResponse(c, http.StatusOK, nil, "")
	// 		return
	// 	}
	// }

	err = h.pushEventsV2(pushEventsRawData)
	if err != nil {
		shared.JSONResponse(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	shared.JSONResponse(c, http.StatusOK, nil, "")
}

func (h *eventsGroup) pushEventsV2(pushEventsRawData []byte) error {
	saveBlockData, err := UnmarshallBlockDataV2(pushEventsRawData)
	if err != nil {
		return err
	}

	err = h.facade.HandlePushEventsV2(*saveBlockData)
	if err != nil {
		return err
	}

	return nil
}

func (h *eventsGroup) revertEvents(c *gin.Context) {
	var revertBlock data.RevertBlock

	err := c.Bind(&revertBlock)
	if err != nil {
		shared.JSONResponse(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	h.facade.HandleRevertEvents(revertBlock)

	shared.JSONResponse(c, http.StatusOK, nil, "")
}

func (h *eventsGroup) finalizedEvents(c *gin.Context) {
	var finalizedBlock data.FinalizedBlock

	err := c.Bind(&finalizedBlock)
	if err != nil {
		shared.JSONResponse(c, http.StatusBadRequest, nil, err.Error())
		return
	}

	h.facade.HandleFinalizedEvents(finalizedBlock)

	shared.JSONResponse(c, http.StatusOK, nil, "")
}

func (h *eventsGroup) createMiddlewares() {
	user, pass := h.facade.GetConnectorUserAndPass()

	if user != "" && pass != "" {
		basicAuth := gin.BasicAuth(gin.Accounts{
			user: pass,
		})
		h.authMiddleware = basicAuth
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *eventsGroup) IsInterfaceNil() bool {
	return h == nil
}
