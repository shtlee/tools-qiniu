package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"qbox.me/api"
	"qbox.me/errcode"
	"strings"
)

// --------------------------------------------------------------------

type Client struct {
	Conf *api.Config
	*http.Client
}

// --------------------------- by RS -----------------------------------------

func (r Client) doPost(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = bodyLength
	return r.Do(req)
}

func doGet(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err = http.DefaultClient.Do(req)
	return resp, err
}

func (r Client) doPostForm(url_ string, data map[string][]string) (resp *http.Response, err error) {
	msg := url.Values(data).Encode()
	return r.doPost(url_, "application/x-www-form-urlencoded", strings.NewReader(msg), (int64)(len(msg)))
}

// --------------------------- by IO -----------------------------------------

func (r Client) doPostByIO(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.Host = r.Conf.HostIp["io_ip"]
	req.ContentLength = bodyLength
	return r.Do(req)
}

// --------------------------- by UP -----------------------------------------

func (r Client) doPostByUP(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.Host = r.Conf.HostIp["up_ip"]
	req.ContentLength = bodyLength
	return r.Do(req)
}

// --------------------------------------------------------------------

type ErrorRet struct {
	Error string "error"
}

func callRet(ret interface{}, resp *http.Response) (code int, err error) {
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		if ret == nil || resp.ContentLength == 0 {
			return
		}
		switch ret.(type) {
		case io.Writer:
			w := ret.(io.Writer)
			io.Copy(w, resp.Body)
			break
		default:
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				code = errcode.UnexceptedResponse
			}
		}
	} else {
		if resp.ContentLength != 0 {
			if ct, ok := resp.Header["Content-Type"]; ok && ct[0] == "application/json" {
				var ret1 ErrorRet
				json.NewDecoder(resp.Body).Decode(&ret1)
				if ret1.Error != "" {
					err = errors.New(ret1.Error)
					return
				}
			}
		}
		err = errcode.Errno(code)
	}
	return
}

// ------------------------------ helpers ------------------------------
func (r Client) CallWithForm(ret interface{}, url string, param map[string][]string) (code int, err error) {
	resp, err := r.doPostForm(url, param)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r Client) CallWith(ret interface{}, url string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {
	resp, err := r.doPost(url, bodyType, body, (int64)(bodyLength))
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r Client) Call(ret interface{}, url string) (code int, err error) {
	resp, err := r.doPost(url, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func Download(url string) (r io.ReadWriter, err error) {
	resp, err := doGet(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	r = new(bytes.Buffer)
	io.Copy(r, resp.Body)
	return r, err
}

// -------------------------- helpers with specified ip ------------------------------------------
func (r Client) CallWithFormBy(clientType string, ret interface{}, url string, param map[string][]string) (code int, err error) {
	resp, err := r.doPostForm(url, param)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r Client) CallWithBy(clientType string, ret interface{}, url string, bodyType string, body io.Reader, bodyLength int64) (code int, err error) {
	resp, err := r.doPost(url, bodyType, body, (int64)(bodyLength))
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func (r Client) CallBy(clientType string, ret interface{}, url string) (code int, err error) {
	resp, err := r.doPost(url, "application/x-www-form-urlencoded", nil, 0)
	if err != nil {
		return errcode.InternalError, err
	}
	return callRet(ret, resp)
}

func DownloadBy(clientType string, url string) (r io.ReadWriter, err error) {
	resp, err := doGet(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	r = new(bytes.Buffer)
	io.Copy(r, resp.Body)
	return r, err
}
