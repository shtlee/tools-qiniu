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

func Upload(rss rs.RSService, entryURI, uploadFile string) error {
	mimeType := "application/octet-stream"
	f, err := os.Open(uploadFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, code, err := rss.Put(entryURI, mimeType, f, fi.Size())
	if code != 200 || err != nil {
		fmt.Println("Upload : ", err, code)
		return err
	}
	return err
}

func CookDomain(domain, regex string) error {
	var normalHost bool

	if normalHost {

	} else {

	}
	return nil
}

func CookDownloadUrl(domain, regex string) error {
	return nil
}

func Publish(rss rs.RSService, domain, bucket string) error {

	return nil
}

func UnPublish(rss rs.RSService, domain string) error {

	return nil
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
	bucketName := "bucket"
	key := "wjl"
	entryURI := bucketName + ":" + key
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
		return
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

	//-----------------------Fetch----------------------------------
	if err := rss.Fetch(getRet.URL, "wjl.jpg"); err != nil {
		fmt.Println("fetch error : ", err)
	}

	domainHost := "wjl.dn.qbox.me"
	//domainIp := "123.123.123.123"
	//-----------------------Publish --------------------------------
	{
		if code, err := rss.Publish(domainHost, "bucket"); code != 200 || err != nil {
			fmt.Println(" |-- rss.Publish --> ", code, err)
			return
		}
	}

	// ----------------------- Get the file published
	{
		downloadUrl := domainHost + "/" + key
		fmt.Println("downloadUrl : ", downloadUrl)

		if err := rss.Fetch(downloadUrl, "wjl2.jpg"); err != nil {
			fmt.Print("Publish Fetch err : ", err)
			return
		}
	}
}
