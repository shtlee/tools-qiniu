package up

import (
    "fmt"
    "os"
    "io"
    "strings"
    "time"
    "crypto/sha1"
    "encoding/hex"
    "net/http"
    "qbox.me/mons/util"
    "qbox.me/auth/digest"
    "qbox.me/service"
    "qbox.me/service/up"
    rss "qbox.me/service/rs"
)

type UpMkBlk struct {
    Name   string `json:name`
    UpHost string `json:"up_host"`
    IoHost string `json:"io_host"`
    RsHost string `json:"rs_host"`
    Ip     string `json:"ip"`

    AccessKey string `json:"access_key"`
    SecretKey string `json:"secret_key"`
    Bucket    string `json:"bucket"`
    DataFile  string `json:"data_file"`

    ChunkSize  int       `json:"chunk_size"`
    BlockBits  uint      `json:"block_bits"`
}


func (self *UpMkBlk) Init(conf []byte) error {
    err := util.LoadConf(self, conf)
    if err != nil {
        return err
    }
    return nil 
}

func (self *UpMkBlk) buildConfig(conf *service.Config, ip string) {
    upIp := make(map[string]string)
    upIp["up_ip"] = ip   // assign to the same ip
    upIp["rs_ip"] = ip
    upIp["io_ip"] = ip 
    conf.HostIp = upIp   // map   string string

    conf.AccessKey = self.AccessKey
    conf.SecretKey = self.SecretKey
    
    host := make(map[string]string)
    host["up"] = self.UpHost
    host["rs"] = self.RsHost
    host["io"] = self.IoHost
    conf.Host = host      // map   string string

    conf.BlockBits = self.BlockBits
    conf.PutChunkSize = self.ChunkSize
}

func (self *UpMkBlk) GenLog(caseName string, begin, end time.Time, duration time.Duration) string {
    sBegin := begin.String()
    msIdx := 23
    sBegin = (string)([]byte(sBegin)[10:msIdx])

    sEnd := end.String()
    sEnd = (string)([]byte(sEnd)[10:msIdx])

    sDuration := duration.String()
    dotIdx := strings.LastIndex(sDuration, ".")
    sDuration = (string)([]byte(sDuration)[:dotIdx+2]) + "ms"
    return fmt.Sprintf("UP ==>> %-18s %-15s %-15s %8s", caseName, sBegin, sEnd, sDuration)
}

func (self *UpMkBlk) doTestMkBlk(ip string) (msg string, err error) {
    var conf service.Config
    // load config
    self.buildConfig(&conf, ip)
    DataFile := self.DataFile
    key := self.Name
    entry := self.Bucket + ":" + key 

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
        checksums []string = make([]string, blockCnt)
        progs []up.BlockProgress = make([]up.BlockProgress, blockCnt)
        ret up.PutRet
        code int
    )
    begin := time.Now() 
    code, err = upservice.Put(f, fi.Size(), checksums, progs, func (int, string) {}, func (int, *up.BlockProgress){})
    
    if err != nil || code != 200 { 
        return 
    }
    code, err = upservice.Mkfile(&ret, "/rs-mkfile/", entry, fi.Size(), "", "", checksums)
    end := time.Now() 
    duration := end.Sub(begin)
    msg = self.GenLog(self.Name + "_resumableput", begin, end, duration)    
    if err != nil || code != 200 { 
        return
    }
    return
}

func (self *UpMkBlk) doRSGet(c service.Config, entryURI string)(msg string, err error) {
    var ret rss.GetRet
    h := sha1.New()

    dt := digest.NewTransport(self.AccessKey, self.SecretKey, nil)
    rsservice, _ := rss.NewRS(&c, dt)

    ret, code, err := rsservice.Get(entryURI, "", "", 3600)
    if err != nil {
        fmt.Println("err : ", err)
        return
    }

    if code != 200 {
        fmt.Println("code : ", err )
        return
    }
    var req *http.Request
    req, err = http.NewRequest("GET", ret.URL, nil)
    if err != nil {
        fmt.Println("err : ", err)
        return
    }
    var resp *http.Response
    resp, err = http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("err : ", err)
        return
    }
    defer resp.Body.Close()
    _, err = io.Copy(h, resp.Body)
    if err != nil {
         fmt.Println("err : ", err)
        return
    }
    hash := hex.EncodeToString(h.Sum(nil))
fmt.Println("hash : ", hash)
    DataSha1 := "f217fcd3958d6888be0f28c7f8e3bd1db176b05b"
    if hash != DataSha1 {
        fmt.Println("sha1 check fails...")
        return
    }
    return 
}


func (self *UpMkBlk) Mon() (msg string, err error) {
        msg1, err := self.doTestMkBlk(self.Ip)
        if err == nil {
            msg += fmt.Sprintln(msg1, " ok")
        } else {
            msg += fmt.Sprintln(msg1, err)
        }
    
    return 
}

