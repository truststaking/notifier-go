package facade

import "errors"

// ErrNilEventsHandler signals that a nil events handler was provided
var ErrNilEventsHandler = errors.New("nil events handler")

// ErrNilWSHandler signals that a nil websocket handler was provided
var ErrNilWSHandler = errors.New("nil websocket handler")

// ErrNilEventsInterceptor signals that a nil events interceptor was provided
var ErrNilEventsInterceptor = errors.New("nil events interceptor")
