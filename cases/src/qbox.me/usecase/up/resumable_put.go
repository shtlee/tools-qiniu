package up

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
    "errors"
	"net/http"
	"os"
	"qbox.me/auth/digest"
	"qbox.me/usecase/util"
	"qbox.me/service"
	rss "qbox.me/service/rs"
	"qbox.me/service/up"
	"time"
)

type UpResuPut struct {
	Name      string `json:name`
	Bucket    string `json:"bucket"`

    Key       string `json:"key"`
	DataFile  string `json:"data_file"`
    DataSha1  string `json:"data_sha1"`
    PutRetryTimes int `json:"put_retry_times"`
    ExpiresTime   int `json:"expires_time"`

	ChunkSize int  `json:"chunk_size"`
	BlockBits uint `json:"block_bits"`

    Url       string 
    Conf      service.Config
    EntryURI  string 
    Env       util.Env
}


func (self *UpResuPut) Init(conf, env []byte) error {
	err := util.LoadConf(self, conf)
    if err != nil {
        return err
    }

    err = util.LoadConf(&self.Env, env)
    if err != nil {
        return err 
    }
	return nil
}

func (self *UpResuPut) buildConfig(conf *service.Config) {
	conf.HostIp = self.Env.Ips // map   string string

	conf.AccessKey = self.Env.AccessKey
	conf.SecretKey = self.Env.SecretKey

	conf.Host = self.Env.Hosts // map   string string

	conf.BlockBits = self.BlockBits
	conf.PutChunkSize = self.ChunkSize
    conf.RPutRetryTimes = self.PutRetryTimes
}

func NewRS(c service.Config) (*rss.RSService, error) {
    dt := digest.NewTransport(c.AccessKey, c.SecretKey, nil)
    return rss.NewRS(&c, dt)
}

func (self *UpResuPut) doTestPut() (msg string, err error) {
	var conf service.Config
	// load config
	self.buildConfig(&conf)
    self.Conf = conf
	DataFile := self.DataFile
	entry := self.Bucket + ":" + self.Key
    self.EntryURI = entry 
	dt := digest.NewTransport(conf.AccessKey, conf.SecretKey, nil)
	upservice, _ := up.NewService(&conf, dt, 1, 1)

	f, err := os.Open(DataFile)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	blockCnt := upservice.BlockCount(fi.Size())

	var (
		checksums []string           = make([]string, blockCnt)
		progs     []up.BlockProgress = make([]up.BlockProgress, blockCnt)
		ret       up.PutRet
		code      int
	)
	begin := time.Now()
	code, err = upservice.Put(f, fi.Size(), checksums, progs, func(int, string) {}, func(int, *up.BlockProgress) {})

	if err != nil || code != 200 {
		return
	}
	code, err = upservice.Mkfile(&ret, "/rs-mkfile/", entry, fi.Size(), "", "", checksums)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestPut", begin, end, duration)
	if err != nil || code != 200 {
		return
	}
	return
}

func (self *UpResuPut) doTestRSGet() (msg string, err error) {
	var ret rss.GetRet
	
	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	rsservice, err := rss.NewRS(&self.Conf, dt)
    if err != nil {
        return 
    }
    begin := time.Now()
	ret, code, err := rsservice.Get(self.EntryURI, "", "", 3600)
    end := time.Now() 
    duration := end.Sub(begin)
    msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestRsGet", begin, end, duration)

	if err != nil || code != 200 {
		return
	}
    self.Url = ret.URL
	return
}

func (self *UpResuPut) doTestDownload()(msg string, err error) {
    h := sha1.New()
    begin := time.Now()
    var req *http.Request
    if req, err = http.NewRequest("GET", self.Url, nil); err != nil {
        return 
    }
    var resp *http.Response
    if resp, err = http.DefaultClient.Do(req); err != nil {
        return 
    }
    defer resp.Body.Close()
    if _, err = io.Copy(h, resp.Body); err != nil {
        return 
    }
    end := time.Now()
    duration := end.Sub(begin)
    msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestDownload", begin, end, duration)

    hash := hex.EncodeToString(h.Sum(nil))
    if hash != self.DataSha1 {
        err = errors.New("check shal failed!")
        return
    }
    return 
}


func (self *UpResuPut) Test() (msg string, err error) {
    msg1 := ""
    msg1, err = self.doTestPut()
    if err == nil {
        msg += fmt.Sprintln(msg1, " ok")
    } else {
        msg += fmt.Sprintln(msg1, err)
        return 
    }

    msg1, err = self.doTestRSGet()
    if err == nil {
        msg += fmt.Sprintln(msg1, " ok")
    } else {
        msg += fmt.Sprintln(msg1, err)
        return 
    }

    msg1, err = self.doTestDownload()
    if err == nil {
        msg += fmt.Sprintln(msg1, " ok")
    } else {
        msg += fmt.Sprintln(msg1, err)
    }
	return
}
