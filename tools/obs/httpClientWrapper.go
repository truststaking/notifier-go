package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("main")

var errInvalidValue = errors.New("invalid value")
var errEmptyUsername = errors.New("empty username provided")
var errEmptyBaseUrl = errors.New("empty url provided")
var errEmptyPassword = errors.New("empty password provided")

const (
	minRequestTimeoutSec = 1
	contentTypeKey       = "Content-Type"
	contentTypeValue     = "application/json"
)

type httpClientWrapper struct {
	httpClient       *http.Client
	useAuthorization bool
	username         string
	password         string
	baseUrl          string
}

// HTTPClientWrapperArgs defines the arguments needed for http client creation
type HTTPClientWrapperArgs struct {
	UseAuthorization  bool
	Username          string
	Password          string
	BaseUrl           string
	RequestTimeoutSec int
}

// NewHTTPWrapperClient creates an instance of httpClient which is a wrapper for http.Client
func NewHTTPWrapperClient(args HTTPClientWrapperArgs) (*httpClientWrapper, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	httpClient.Timeout = time.Duration(args.RequestTimeoutSec) * time.Second

	return &httpClientWrapper{
		httpClient:       httpClient,
		useAuthorization: args.UseAuthorization,
		username:         args.Username,
		password:         args.Password,
		baseUrl:          args.BaseUrl,
	}, nil
}

func checkArgs(args HTTPClientWrapperArgs) error {
	if args.BaseUrl == "" {
		return errEmptyBaseUrl
	}
	if args.RequestTimeoutSec < minRequestTimeoutSec {
		return fmt.Errorf("%w, provided: %v, minimum: %v", errInvalidValue, args.RequestTimeoutSec, minRequestTimeoutSec)
	}
	if args.UseAuthorization {
		if args.Username == "" {
			return errEmptyUsername
		}
		if args.Password == "" {
			return errEmptyPassword
		}
	}

	return nil
}

// Post can be used to send POST requests. It handles marshalling to/from json
func (h *httpClientWrapper) Post(
	route string,
	payload interface{},
) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s", h.baseUrl, route)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set(contentTypeKey, contentTypeValue)

	if h.useAuthorization {
		req.SetBasicAuth(h.username, h.password)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			bodyCloseErr := resp.Body.Close()
			if bodyCloseErr != nil {
				log.Warn("error while trying to close response body", "err", bodyCloseErr.Error())
			}
		}
	}()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn("httpClient: received HTTP status", "code", resp.StatusCode, "responseBody", string(resBody))
		return fmt.Errorf("HTTP status code: %d, %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *httpClientWrapper) IsInterfaceNil() bool {
	return h == nil
}
