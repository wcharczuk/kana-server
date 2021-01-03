package r2

import (
	"time"
)

// OptTLSHandshakeTimeout sets the client transport TLSHandshakeTimeout.
func OptTLSHandshakeTimeout(d time.Duration) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		transport.TLSHandshakeTimeout = d
		return nil
	}
}
