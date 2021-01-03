package web

import "net/http"

// PanicHandler is a handler for panics that also takes an error.
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})
