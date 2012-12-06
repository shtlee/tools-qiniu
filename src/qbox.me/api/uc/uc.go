package uc

import (
	"errors"
	"net/http"
	. "qbox.me/api"
	"qbox.me/rpc"
	"strconv"
)

type UCService struct {
	*Config
	Conn rpc.Client
}

func New(c *Config, t http.RoundTripper) (s *UCService, err error) {
	if c == nil {
		err = errors.New("Must have a config file")
		return
	}
	client := &http.Client{Transport: t}
	s = &UCService{c, rpc.Client{c, client}}
	return
}

func (s *UCService) AntiLeechMode(bucket string, mode int) (code int, err error) {
	param := map[string][]string{
		"bucket": {bucket},
		"mode":   {strconv.Itoa(mode)},
	}
	url := s.Host["uc"] + "/antiLeechMode"
	return s.Conn.CallWithForm(nil, url, param)
}

func (s *UCService) AddAntiLeech(bucket string, mode int, pattern string) (code int, err error) {
	param := map[string][]string{
		"bucket":  {bucket},
		"mode":    {strconv.Itoa(mode)},
		"action":  {"add"},
		"pattern": {pattern},
	}
	url := s.Host["uc"] + "/referAntiLeech"
	return s.Conn.CallWithForm(nil, url, param)
}

func (s *UCService) CleanCache(bucket string) (code int, err error) {
	param := map[string][]string{
		"bucket": {bucket},
	}
	url := s.Host["uc"] + "/refreshBucket"
	return s.Conn.CallWithForm(nil, url, param)
}

func (s *UCService) DelAntiLeech(bucket string, mode int, pattern string) (code int, err error) {
	param := map[string][]string{
		"bucket":  {bucket},
		"mode":    {strconv.Itoa(mode)},
		"action":  {"del"},
		"pattern": {pattern},
	}
	url := s.Host["uc"] + "/referAntiLeech"
	return s.Conn.CallWithForm(nil, url, param)
}
