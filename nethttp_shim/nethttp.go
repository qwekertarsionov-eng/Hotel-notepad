// Package nethttp is a compatibility shim so existing tests that import
// "nethttp" (an alias for net/http) compile without modification.
package nethttp

import "net/http"

type Request = http.Request
type Header = http.Header

var NewRequest = http.NewRequest
var MethodPost = http.MethodPost
