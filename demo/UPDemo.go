package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	//"net/http"
	"os"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/up"
	"qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
	//"qbox.me/rpc"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("checkErr : ", err)
		os.Exit(-1)
	}
}

func LoadConf(fname string, conf interface{}) (err error) {
	b, err := ioutil.ReadFile(fname)
	checkError(err)

	err = json.Unmarshal(b, &conf)
	checkError(err)
	return nil
}

func NewUP(config api.Config) (*up.Service, error) {
	auth := &uptoken.AuthPolicy{}
	tokenString := uptoken.MakeAuthTokenString(config.AccessKey, config.SecretKey, auth)
	upTransport := uptoken.NewTransport(tokenString, nil)
	return up.New(&config, upTransport)
}

func NewRS(c api.Config) (*rs.RSService, error) {
	dt := digest.NewTransport(c.AccessKey, c.SecretKey, nil)
	return rs.New(&c, dt)
}

func main() {
	var conf api.Config
	// load config
	LoadConf("qbox.conf", &conf)
	rss, err := NewRS(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	entryURI := "bucket" + ":" + "wjl"
	mimeType := "application/octet-stream"
	testFile := "mm.jpg"
	f, err := os.Open(testFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
	}
	ret, code, err := rss.Put(entryURI, mimeType, f, fi.Size())
	if code/100 != 2 {
		fmt.Println(code)
	}
	fmt.Println(" ** hash : ", ret.Hash)

	//--------------------------------------------------------
	getRet, code, err := rss.Get(entryURI, "", "", 3600)
	if err != nil || code != 200 {
		fmt.Println("rss.Get : ", err, code)
	}
	fmt.Println(" ** url       : ", getRet.URL)
	fmt.Println(" ** hash      : ", getRet.Hash)
	fmt.Println(" ** mimetype  : ", getRet.MimeType)
	fmt.Println(" ** fsize     : ", getRet.Fsize)
	fmt.Println(" ** expiry    : ", getRet.Expiry)

	//-----------------------download----------------------------------

	if err := rss.Fetch(getRet.URL, "wjl.jpg"); err != nil {
		fmt.Println("fetch error : ", err)
	}
}
