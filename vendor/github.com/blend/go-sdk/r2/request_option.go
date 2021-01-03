package r2

import "net/http"

// RequestOption translates a webutil.RequestOption to a r2.Option.
func RequestOption(opt func(*http.Request) error) Option {
	return func(r *Request) error {
		return opt(r.Request)
	}
}
