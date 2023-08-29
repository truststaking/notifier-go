package gin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	apiErrors "github.com/multiversx/mx-chain-notifier-go/api/errors"
	"github.com/multiversx/mx-chain-notifier-go/api/groups"
	"github.com/multiversx/mx-chain-notifier-go/api/shared"
	"github.com/multiversx/mx-chain-notifier-go/common"
	"github.com/multiversx/mx-chain-notifier-go/config"
)

const (
	defaultRestInterface = "localhost:5000"
)

var log = logger.GetOrCreate("api/gin")

// ArgsWebServerHandler holds the arguments needed to create a web server handler
type ArgsWebServerHandler struct {
	Facade  shared.FacadeHandler
	Configs config.Configs
}

// webServer is a wrapper for gin.Engine, holding additional components
type webServer struct {
	sync.RWMutex
	facade       shared.FacadeHandler
	httpServer   shared.HTTPServerCloser
	configs      config.Configs
	groups       map[string]shared.GroupHandler
	wasTriggered bool
	cancelFunc   func()
}

// NewWebServerHandler creates and configures an instance of webServer
func NewWebServerHandler(args ArgsWebServerHandler) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &webServer{
		facade:       args.Facade,
		configs:      args.Configs,
		groups:       make(map[string]shared.GroupHandler),
		wasTriggered: false,
	}, nil
}

func checkArgs(args ArgsWebServerHandler) error {
	if check.IfNil(args.Facade) {
		return apiErrors.ErrNilFacadeHandler
	}
	if args.Configs.Flags.APIType == "" {
		return common.ErrInvalidAPIType
	}

	return nil
}

func (w *webServer) getWSAddr() string {
	addr := w.configs.GeneralConfig.ConnectorApi.Host
	if addr == "" {
		return defaultRestInterface
	}

	if !strings.Contains(addr, ":") {
		return fmt.Sprintf(":%s", addr)
	}

	return addr
}

// Run starts the server and the Hub as goroutines
// It returns an instance of http.Server
func (w *webServer) Run() error {
	w.Lock()
	defer w.Unlock()

	var err error

	if w.wasTriggered == true {
		log.Error("Web server has been already triggered successfuly once")
		return nil
	}

	engine := gin.New()
	engine.Use(cors.Default())
	engine.Use(gin.Recovery())
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	err = w.createGroups()
	if err != nil {
		return err
	}

	w.registerRoutes(engine)

	addr := w.getWSAddr()

	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	w.httpServer, err = NewHTTPServerWrapper(server)
	if err != nil {
		return err
	}

	go w.httpServer.Start()

	w.wasTriggered = true

	return nil
}

func (w *webServer) createGroups() error {
	groupsMap := make(map[string]shared.GroupHandler)

	eventsGroup, err := groups.NewEventsGroup(w.facade)
	if err != nil {
		return err
	}
	groupsMap["events"] = eventsGroup

	statusGroup, err := groups.NewStatusGroup(w.facade)
	if err != nil {
		return err
	}
	groupsMap["status"] = statusGroup

	if w.configs.Flags.APIType == common.WSAPIType {
		hubHandler, err := groups.NewHubGroup(w.facade)
		if err != nil {
			return err
		}
		groupsMap["hub"] = hubHandler
	}

	w.groups = groupsMap

	return nil
}

func (w *webServer) registerRoutes(ginEngine *gin.Engine) {
	for groupName, groupHandler := range w.groups {
		log.Info("registering API group", "group name", groupName)

		ginGroup := ginEngine.Group(fmt.Sprintf("/%s", groupName))

		groupHandler.RegisterRoutes(ginGroup, w.configs.ApiRoutesConfig)
	}
}

// Close will handle the closing of inner components
func (w *webServer) Close() error {
	if w.cancelFunc != nil {
		w.cancelFunc()
	}

	w.Lock()
	err := w.httpServer.Close()
	w.Unlock()

	if err != nil {
		err = fmt.Errorf("%w while closing the http server in gin/webServer", err)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (w *webServer) IsInterfaceNil() bool {
	return w == nil
}
