package fop

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"qbox.me/service/rs"
	da "qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
	"qbox.me/usecase/util"
	"time"
	"qbox.me/service"
    "os"
)

type FopImgOp struct {
	Name 			string 			`json:"name"`
	BucketName  	string          `json:"bucket"`
	Key 			string          `json:"key"`
	
    ChunkSize 		int  			`json:"chunk_size"`
	BlockBits 		uint 			`json:"block_bits"`
	
    SrcImg      	string         	`json:"source_file"`  
	TargetImg   	string      	`json:"target_file"`
    Op              string          `json:"op"`

	Fopd        	string          `json:"fopd"`
	FopdLogFile 	string          `json:"log_fopd"`

	Conf        	service.Config
	Env         	util.Env 
}

func (self *FopImgOp) buildConfig(conf *service.Config) {
	conf.HostIp = self.Env.Ips // map   string string

	conf.AccessKey = self.Env.AccessKey
	conf.SecretKey = self.Env.SecretKey

	conf.Host = self.Env.Hosts // map   string string

	conf.BlockBits = self.BlockBits
	conf.PutChunkSize = self.ChunkSize
}

func (self *FopImgOp) Init(conf, env []byte) error {
	if err := util.LoadConf(self, conf); err != nil {
		return err 
	}
	
	if err := util.LoadConf(&self.Env, env); err != nil {
		return err 
	}

	var config service.Config
	self.buildConfig(&config)
	self.Conf = config 
	return nil
}

// upload the file and get the download url 
func (self *FopImgOp) doTestGetImgUrl() (url string, err error) {
	entry := self.BucketName + ":" + self.Key

	dt := da.NewTransport(self.Conf.AccessKey, self.Conf.SecretKey, nil)
    rsservice, err := rs.NewRS(&self.Conf, dt)
    if err != nil {
    	return 
    }
    authPolicy := &uptoken.AuthPolicy{
            Scope:    entry,
            Deadline: 3600,
    }
    authPolicy.Deadline += uint32(time.Now().Unix())
    token := uptoken.MakeAuthTokenString(self.Conf.AccessKey, self.Conf.SecretKey, authPolicy)
    _, code, err := rsservice.Upload(entry, self.SrcImg, "", "", "", token)

    if err != nil || code != 200 {
    	return 
    }
    /*f, err := os.Open(self.SrcImg)
    if err != nil {
    	return 
    }
    defer f.Close()
    fi, err := f.Stat()
    if err != nil {
    	return
    }

   	_, code, err := rsservice.Put(entry, "", f, fi.Size())
 fmt.Println("rs.Put ", code, err )
   	if err != nil || code != 200 {
   		return 
   	}*/

	getRet, code, err := rsservice.Get(entry, "", "", 3600)
	if err != nil || code != 200 {
		return
	}
	
	url = getRet.URL
	return
}

func (self *FopImgOp) doTestImgOp(downloadUrl string) (msg string, err error) {
	begin := time.Now()	
	url := downloadUrl + "?" + self.Op
	netBuf, err := util.DoHttpGet(url)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("Fp    "+self.Env.Id+"_"+self.Name+"_doTestImgOp", begin, end, duration)
	if err != nil {
		return
	}
    targetFile, err := os.Open(self.TargetImg)
    if err != nil {
        return 
    }
    _, err = util.CheckImg(netBuf, targetFile)
    if err != nil {
        return 
    }

	return
}

func (self *FopImgOp) Test() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgUrl()
	if err != nil {
		return 
	}

	msg1, err = self.doTestImgOp(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	return
}
