package main 

import (
    "fmt"
    "qbox.me/api/up"
    "qbox.me/api"
    "qbox.me/api/rs"
    "io/ioutil"
    "net/http"
    "os"
    gio "io"
    "crypto/sha1"
    rss "qbox.me/api/rs"
    "encoding/hex"
    "encoding/json"
    "qbox.me/auth/digest"
)

func NewRS(c api.Config) (*rs.RSService, error) {
    dt := digest.NewTransport(c.AccessKey, c.SecretKey, nil)
    return rs.New(&c, dt)
}

func LoadConf(fname string, conf interface{}) (err error) {
    b, err := ioutil.ReadFile(fname)
    if err != nil {
        fmt.Println(err)
        return 
    }
    err = json.Unmarshal(b, &conf)
    if err != nil {
        fmt.Println(err)
        return 
    }
    return nil
}

func main() {
     var conf api.Config
    // load config
    LoadConf("qbox.conf", &conf)
     
    DataFile := "qbox.conf"
    key := "a.txt"
    entry := "bucket" + ":" + key 

    dt := digest.NewTransport(conf.AccessKey, conf.SecretKey, nil)
    upservice, _ := up.NewService(&conf, dt, 1, 1)

    
    f, err := os.Open(DataFile)
    fi, err := f.Stat()
    blockCnt := up.BlockCount(fi.Size())

    var checksums []string = make([]string, blockCnt)
    var progs []up.BlockProgress = make([]up.BlockProgress, blockCnt)

    // put resumable
    {
        var ret up.PutRet
        var code int 
        code, err = upservice.Put(f, fi.Size(), checksums, progs, func (int, string) {}, func (int, *up.BlockProgress){})
        
        if err != nil || code != 200 { 
           fmt.Println(err, code)
            return
        }

        code, err = upservice.Mkfile(&ret, "/rs-mkfile/", entry, fi.Size(), "", "", checksums)
        if err != nil || code != 200 { 
           fmt.Println(err, code)
            return
        }
    }

    f.Close()

fmt.Println("download.............")
   
    rs2, err := NewRS(conf)
    if err != nil {
        fmt.Println("err : ", err)
        return
    }
  
    // download
    {
        var ret rss.GetRet
        h := sha1.New()
        code := 200
        ret, code, err = rs2.Get(entry, "", "", 3600)
fmt.Println("ret : ", ret )
fmt.Println("URL : ", ret.URL)
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
        _, err = gio.Copy(h, resp.Body)
        if err != nil {
             fmt.Println("err : ", err)
            return
        }
        hash := hex.EncodeToString(h.Sum(nil))
fmt.Println("hash : ", hash)
        DataSha1 := "aec60d39ff5810f4e4d4fc3373ba58a22803a62d"
        if hash != DataSha1 {
            fmt.Println("sha1 check fails...")
            return
        }

        rs2.Fetch(ret.URL, "aa.confif")
    }

}
