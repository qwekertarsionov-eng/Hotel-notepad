// Package httptest is a compatibility shim for "nethttp/httptest".
package httptest

import "net/http/httptest"

type ResponseRecorder = httptest.ResponseRecorder
type Server = httptest.Server

var NewRecorder = httptest.NewRecorder
