package httputil

import (
	"net/http"
	"qbox.me/service"
)

type RsClient struct {
	Conf *service.Config
	*http.Client
}

/*
func (r Client) doPostByRS(url string, bodyType string, body io.Reader, bodyLength int64) (resp *http.Response, err error) {
	fmt.Println(" |- doPostByRS --> ", url)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", bodyType)
	req.Host = r.Conf.Host["rs"]
	fmt.Println("req.Host --> ", req.Host)
	req.ContentLength = bodyLength
	return r.Do(req)
}

func doGetByRS(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err = http.DefaultClient.Do(req)
	return resp, err
}

func (r Client) doPostFormByRS(url_ string, data map[string][]string) (resp *http.Response, err error) {
	msg := url.Values(data).Encode()
	return r.doPostByRS(url_, "application/x-www-form-urlencoded", strings.NewReader(msg), (int64)(len(msg)))
}
*/
