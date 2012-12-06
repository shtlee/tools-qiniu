package rpc

import (
	"io"
	"net/http"
)

type Http interface {
	DoPost(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error)

	DoGet(url string) (resp *http.Response, err error)

	DoPostForm(url_ string, data map[string][]string) (resp *http.Response, err error)
}

