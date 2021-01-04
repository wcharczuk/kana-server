package r2

import "github.com/blend/go-sdk/webutil"

// OptXMLBody sets the post body on the request.
func OptXMLBody(obj interface{}) Option {
	return RequestOption(webutil.OptXMLBody(obj))
}
