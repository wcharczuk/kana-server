package r2

import "github.com/blend/go-sdk/webutil"

// OptJSONBody sets the post body on the request.
func OptJSONBody(obj interface{}) Option {
	return RequestOption(webutil.OptJSONBody(obj))
}
