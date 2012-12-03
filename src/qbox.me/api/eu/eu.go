package eu

import (
	"errors"
	"net/http"
	. "qbox.me/api"
	"qbox.me/rpc"
	"strconv"
)

type EUService struct {
	*Config
	Conn rpc.Client
}

func New(c *Config, t http.RoundTripper) (s *EUService, err error) {
	if c == nil {
		err = errors.New("Must have a config file")
		return
	}
	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = &EUService{c, rpc.Client{c, client}}
	return
}

type Watermark struct {
	Font     string `json:"font"`
	Fill     string `json:"fill"`
	Text     string `json:"text"`
	Bucket   string `json:"bucket"`
	Dissolve string `json:"dissolve"`
	Gravity  string `json:"gravity"`
	FontSize int    `json:"fontsize"` // 0 表示默认。单位: 缇，等于 1/20 磅
	Dx       int    `json:"dx"`
	Dy       int    `json:"dy"`
}

func (s *EUService) GetWatermark(customer string) (ret Watermark, code int, err error) {

	params := map[string][]string{
		"customer": {customer},
	}
	code, err = s.Conn.CallWithForm(&ret, s.Host["eu"]+"/wmget", params)
	return
}

func (s *EUService) SetWatermark(customer string, args *Watermark) (code int, err error) {

	params := map[string][]string{
		"text": {args.Text},
		"dx":   {strconv.Itoa(args.Dx)},
		"dy":   {strconv.Itoa(args.Dy)},
	}
	if customer != "" {
		params["customer"] = []string{customer}
	}
	if args.Font != "" {
		params["font"] = []string{args.Font}
	}
	if args.FontSize != 0 {
		params["fontsize"] = []string{strconv.Itoa(args.FontSize)}
	}
	if args.Fill != "" {
		params["fill"] = []string{args.Fill}
	}
	if args.Bucket != "" {
		params["bucket"] = []string{args.Bucket}
	}
	if args.Dissolve != "" {
		params["dissolve"] = []string{args.Dissolve}
	}
	if args.Gravity != "" {
		params["gravity"] = []string{args.Gravity}
	}
	return s.Conn.CallWithForm(nil, s.Host["eu"]+"/wmset", params)
}
