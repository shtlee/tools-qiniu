package rs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	. "qbox.me/api"
	"qbox.me/rpc"
	"strconv"
	"time"
	"strings"
)

type RSService struct {
	*Config
	Conn rpc.Client
}

func NewRS(c *Config, t http.RoundTripper) (s *RSService, err error) {
	if c == nil {
		err = errors.New("Must have a config file")
		return
	}
	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = &RSService{c, rpc.Client{c, client}}
	return
}

type PutRet struct {
	Hash string `json:"hash"`
}

type GetRet struct {
	URL      string `json:"url"`
	Hash     string `json:"hash"`
	MimeType string `json:"mimeType"`
	Fsize    int64  `json:"fsize"`
	Expiry   int64  `json:"expires"`
}

type Entry struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
}

func (s *RSService) Put(
	entryURI, mimeType string, body io.Reader, bodyLength int64) (ret PutRet, code int, err error) {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	url := s.HostIp["io_ip"] + "/rs-put/" + EncodeURI(entryURI) + "/mimeType/" + EncodeURI(mimeType)
	code, err = s.Conn.CallWithBy("io", &ret, url, "application/octet-stream", body, (int64)(bodyLength))
	return
}

// 动态获取文件授权后的临时下载链接
func (s *RSService) Get(entryURI, base, attName string, expires int) (data GetRet, code int, err error) {
	url := s.HostIp["rs_ip"] + "/get/" + EncodeURI(entryURI)
	if base != "" {
		url += "/base/" + base
	}
	if attName != "" {
		url += "/attName/" + EncodeURI(attName)
	}
	if expires > 0 {
		url += "/expires/" + strconv.Itoa(expires)
	}
	//code, err = s.Conn.Call(&data, url)
	code, err = s.Conn.CallBy("rs", &data, url)
	if code == 200 {
		data.Expiry += time.Now().Unix()
	}
	return
}

// 
func replaceHostWithIP(url, host, ip string) string {
	return strings.Replace(url, host, ip, 1)
}

// Fetch  downloads a file specified the url and then stores it as the fname
// on the disk.
func (s *RSService) Fetch(url, saveAs string) error {
	fi, err := os.OpenFile(saveAs, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	defer fi.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	ip := string([]byte(s.Config.HostIp["io_ip"][7:]))
	url = replaceHostWithIP(url, s.Config.Host["io"], ip)
	reader, err := s.Conn.DownloadBy("io", url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	io.Copy(fi, reader)
	return err
}

func (s *RSService) Stat(entryURI string) (entry Entry, code int, err error) {
	code, err = s.Conn.Call(&entry, s.Host["rs"]+"/stat/"+EncodeURI(entryURI))
	return
}

func (s *RSService) Delete(entryURI string) (code int, err error) {
	return s.Conn.Call(nil, "http://"+s.Host["rs"]+"/delete/"+EncodeURI(entryURI))
}

func (s *RSService) Mkbucket(bucketname string) (code int, err error) {
	return s.Conn.Call(nil, s.Host["rs"]+"/mkbucket/"+bucketname)
}

func (s *RSService) Drop(entryURI string) (code int, err error) {
	return s.Conn.Call(nil, s.Host["rs"]+"/drop/"+entryURI)
}

func (s *RSService) Move(entryURISrc, entryURIDest string) (code int, err error) {
	return s.Conn.Call(nil, s.Host["rs"]+"/move/"+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
}

func (s *RSService) Copy(entryURISrc, entryURIDest string) (code int, err error) {
	return s.Conn.Call(nil, s.Host["rs"]+"/copy/"+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
}

func (s *RSService) Publish(domain, table string) (code int, err error) {
	fmt.Println(" |--- rs Publish --> ", domain, table)
	//return s.Conn.Call(nil, "http://"+s.Host["rs"]+"/publish/"+EncodeURI(domain)+"/from/"+table)
	return s.Conn.CallBy("rs", nil, s.HostIp["rs_ip"]+"/publish/"+EncodeURI(domain)+"/from/"+table)
}

func (s *RSService) Unpublish(domain string) (code int, err error) {
	return s.Conn.Call(nil, s.Host["rs"]+"/unpublish/"+EncodeURI(domain))
}

// -------------------Batcher to do -----------------------------------

type BatchRet struct {
	Data  interface{} `json:"data"`
	Code  int         `json:"code"`
	Error string      `json:"error"`
}

type Batcher struct {
	s1  *RSService
	op  []string
	ret []BatchRet
}

func (s *RSService) NewBatcher() *Batcher {
	return &Batcher{s1: s}
}

func (b *Batcher) operate(entryURI string, method string) {
	b.op = append(b.op, method+EncodeURI(entryURI))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) operate2(entryURISrc, entryURIDest string, method string) {
	b.op = append(b.op, method+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) Stat(entryURI string) {
	b.operate(entryURI, "/stat/")
}

func (b *Batcher) Get(entryURI string) {
	b.operate(entryURI, "/get/")
}

func (b *Batcher) Delete(entryURI string) {
	b.operate(entryURI, "/delete/")
}

func (b *Batcher) Move(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/move/")
}

func (b *Batcher) Copy(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/copy/")
}

func (b *Batcher) Reset() {
	b.op = nil
	b.ret = nil
}

func (b *Batcher) Len() int {
	return len(b.op)
}

func (b *Batcher) Do() (ret []BatchRet, code int, err error) {
	s := b.s1
	code, err = s.Conn.CallWithForm(&b.ret, s.Host["rs"]+"/batch", map[string][]string{"op": b.op})
	ret = b.ret
	return
}

func (s RSService) Upload(entryURI, localFile, mimeType, customMeta, callbackParam string, upToken string) (ret PutRet, code int, err error) {

	return s.UploadEx(upToken, localFile, entryURI, mimeType, customMeta, callbackParam, -1, -1)
}

func (s RSService) UploadEx(upToken string, localFile, entryURI string, mimeType, customMeta, callbackParam string,
	crc int64, rotate int) (ret PutRet, code int, err error) {

	action := "/rs-put/" + rpc.EncodeURI(entryURI)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	action += "/mimeType/" + rpc.EncodeURI(mimeType)
	if customMeta != "" {
		action += "/meta/" + rpc.EncodeURI(customMeta)
	}
	if crc >= 0 {
		action += "/crc32/" + strconv.FormatInt(crc, 10)
	}
	if rotate >= 0 {
		action += "/rotate/" + strconv.FormatInt(int64(rotate), 10)
	}
	url := "http://up.qbox.me" + "/upload"

	multiParams := map[string][]string{
		"action": {action},
		"file":   {"@" + localFile},
		"auth":   {upToken},
	}
	if callbackParam != "" {
		multiParams["params"] = []string{callbackParam}
	}

	code, err = s.Conn.CallWithMultipart(&ret, url, multiParams)
	return
}