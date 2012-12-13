package fop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	//	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	//"math/rand"
	"net/http"
	"os"
	//"os/exec"
	//"math"
	"strings"
	//gup "qbox.us/api/up"
	"encoding/base64"
	"qbox.me/service/rs"
	da "qbox.me/auth/digest"
	"errors"
	"qbox.me/usecase/util"
	"strconv"
	//"time"
	"qbox.me/log"
	"qbox.me/sstore"
	//	"sync"
	"time"
	"qbox.me/service"
)

type Fileop struct {
	// Ops []string `json:"ops"`
	Name 		string 			`json:"name"`
	Hosts       map[string]string 	`json:"hosts"`
	Ips         map[string]string 	`json:"ips"`

	ImageFile   string          `json:"image_file"`
	BucketName  string          `json:"bucket"`
	AccessKey   string          `json:"access_key"`
	SecretKey   string          `json:"secret_key"`
	ImgExif     ImageExif       `json:"img_exif"`

	ChunkSize 	int  			`json:"chunk_size"`
	BlockBits 	uint 			`json:"block_bits"`

	ImgInfo     ImageInfo       `json:"img_info"`
	FileopReq   []FileopRequest `json:"fileop_req"`
	Fopd        string          `json:"fopd"`
	FopdLogFile string          `json:"log_fopd"`
	FopdLoger   *log.Logger
	Conf        service.Config
}

type ImageExif struct {
	ExifJson string `json:"exif"`
}

type ImageInfo struct {
	Format     string `json:"format"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	ColorModel string `json:"colorModel"`
}

type FileopRequest struct {
	Name       string `json:name`
	TargetFile string `json:"target_file"`
	Op         string `json:"op"`
}

const KeyHintESTest = 139
const KeyHintMockFS = 113
const KeyHintFSTest = 103
const KeyHintIOTest = 104
const KeyHintRSTest = 105
const KeyHintPubTest = 106

var KeyMockFS = []byte("qbox.mockfs")
var KeyFSTest = []byte("qbox.fs.test")
var KeyIOTest = []byte("qbox.io.test")
var KeyESTest = []byte("qbox.es.test")
var KeyRSTest = []byte("qbox.rs.test")
var KeyPubTest = []byte("qbox.pub.test")
var KeyFinder = sstore.SimpleKeyFinder(map[uint32][]byte{
	KeyHintMockFS:  KeyMockFS,
	KeyHintFSTest:  KeyFSTest,
	KeyHintIOTest:  KeyIOTest,
	KeyHintESTest:  KeyESTest,
	KeyHintRSTest:  KeyRSTest,
	KeyHintPubTest: KeyPubTest,
})

func decodeFh(efh string) *sstore.FhandleInfo {
	return sstore.DecodeFhandle(efh, "", KeyFinder)
}

func (self *Fileop) buildConfig(conf *service.Config) {
	conf.HostIp = self.Ips // map   string string

	conf.AccessKey = self.AccessKey
	conf.SecretKey = self.SecretKey

	conf.Host = self.Hosts // map   string string

	conf.BlockBits = self.BlockBits
	conf.PutChunkSize = self.ChunkSize
}

func (self *Fileop) Init(conf []byte) error {
	err := util.LoadConf(self, conf)
	if err != nil {
		return err
	}
	fi, err := os.OpenFile(self.FopdLogFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	self.FopdLoger = log.New(fi, "", 0)
	return nil
}

func (self *Fileop) doTestGetImgDownloadUrl() (url string, err error) {
	var conf service.Config
	// load config
	self.buildConfig(&conf)
	self.Conf = conf 
	fmt.Println("conf : ", conf)
	//rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	//entry := fp.BucketName + ":" + fp.ImageFile + "_" + strconv.FormatInt(rnd.Int63(), 10)
	entry := self.BucketName + ":fixed_entry_not_delete"
	/*policy := &gup.AuthPolicy{
		Scope:    entry,
		Deadline: 3600,
		Customer: "1234",
	}
	policy.Deadline += uint32(time.Now().Unix())*/
	//token := gup.MakeAuthTokenString(fp.AccessKey, fp.SecretKey, policy)

	/*_, code, err := rss.Upload(entry, fp.ImageFile, "", "", "", token)

	if err != nil || code != 200 {
		return
	}*/


	dt := da.NewTransport(self.Conf.AccessKey, self.Conf.SecretKey, nil)
    conn, err := rs.NewRS(&self.Conf, dt)
    if err != nil {
    	return 
    }

	begin := time.Now()
	getRet, code, err := conn.Get(entry, "", "", 3600)

	if err != nil || code != 200 {
		return
	}
	end := time.Now()
	self.DoLog2("doTestImgUrl", begin, end, end.Sub(begin))

	url = getRet.URL
	return
}

func (self *Fileop) doTestGetImgInfo(downloadUrl string) (err error) {
	url := downloadUrl + "imageInfo"
	netBuf, err := self.DoHttpGet(url, "doTestImgInfo")
	if err != nil {
		return
	}

	var serImgInfo ImageInfo
	json.Unmarshal(netBuf.Bytes(), &serImgInfo)
	locImgInfo := self.ImgInfo

	if err = checkImgInfo(serImgInfo, locImgInfo); err != nil {
		return
	}
	return
}

func (self *Fileop) doTestFileopReq(url string, fileopReq FileopRequest) (err error) {
	url = url + fileopReq.Op
	b, err := self.DoHttpGet(url, fileopReq.Name)
	if err != nil {
		return
	}
	fi, err := os.Open(fileopReq.TargetFile)
	if err != nil {
		return
	}
	defer fi.Close()
	ok, err := self.checkFileop(b, fi)
	if !ok || err != nil {
		return
	}
	return
}

func (self *Fileop) checkFileop(s, t io.Reader) (bool, error) {
	code, err := checkImg(s, t)
	return code == 0, err
}

func (self *Fileop) doTestGetImgEXIF(downloadUrl string) (err error) {
	url := downloadUrl + "exif"
	netBuf, err := self.DoHttpGet(url, "doTestImgExif")
	if err != nil {
		return
	}
	var exif ImageExif
	json.Unmarshal(netBuf.Bytes(), &exif)
	if err = checkImgExif(self.ImgExif.ExifJson, exif.ExifJson); err != nil {
		return
	}
	return
}

func (self *Fileop) DoLog(msg string) {
	self.FopdLoger.Println(msg)
}

func (self *Fileop) DoLog2(caseName string, begin, end time.Time, duration time.Duration) {
	sBegin := begin.String()
	msIdx := 23
	sBegin = (string)([]byte(sBegin)[10:msIdx])

	sEnd := end.String()
	sEnd = (string)([]byte(sEnd)[10:msIdx])

	sDuration := duration.String()
	dotIdx := strings.LastIndex(sDuration, ".")
	sDuration = (string)([]byte(sDuration)[:dotIdx+2]) + "ms"
	self.FopdLoger.Printf("Fileop ==>>%-18s %-15s %-15s %8s", caseName, sBegin, sEnd, sDuration)
}

func (self *Fileop) DoHttpGet(url, caseName string) (b *bytes.Buffer, err error) {
	begin := time.Now()
	var (
		req  *http.Request
		resp *http.Response
	)

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}

	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()
	end := time.Now()
	self.DoLog2(caseName, begin, end, end.Sub(begin))
	b = new(bytes.Buffer)
	io.Copy(b, resp.Body)
	return
}
func approachTo(a1 uint32, a2 uint32) bool {
	var max, d int
	if a1 < a2 {
		max = int(a2)
		d = int(a2 - a1)
	} else {
		max = int(a1)
		d = int(a1 - a2)
	}
	if d <= max/10 {
		return true
	}
	return false
}

func checkImg(src, tgt io.Reader) (int, error) {
	image1, format1, err1 := image.Decode(src)
	image2, format2, err2 := image.Decode(tgt)
	if err1 != nil || err2 != nil {
		return 1, errors.New(fmt.Sprintf("From Servier error : %v , From Local error : %v \n", err1, err2))
	}
	if format2 != format1 {
		return 2, errors.New("Unmatched format!")
	}
	if image1.Bounds() != image2.Bounds() {
		return 3, errors.New("Unmatched bounds")
	}

	total, miss := 0, 0
	for i := image1.Bounds().Min.X; i <= image1.Bounds().Max.X; i++ {
		for j := image1.Bounds().Min.Y; j <= image1.Bounds().Max.Y; j++ {
			total++
			r1, g1, b1, a1 := image1.At(i, j).RGBA()
			r2, g2, b2, a2 := image2.At(i, j).RGBA()
			if !approachTo(r1, r2) || !approachTo(g1, g2) || !approachTo(b1, b2) || !approachTo(a1, a2) {
				return 4, errors.New("Differs two much!")
			}
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				miss++
			}
		}
	}
	if miss > total/20 {
		return 4, errors.New("Missed too much!")
	}
	return 0, nil
}

func checkImgExif(locExif, serverExif string) error {
	if locExif != serverExif {
		return errors.New("Unmatched exif!")
	}
	return nil
}

func checkImgInfo(server ImageInfo, local ImageInfo) error {
	switch {
	case server.Height != local.Height:
		return errors.New(fmt.Sprintf("Unmatched height! Expect : %d, But get %d \n",
			local.Height, server.Height))
	case server.Width != local.Width:
		return errors.New(fmt.Sprintf("Unmatched width! Expect : %d, But get %d \n",
			local.Width, server.Width))
	case server.ColorModel != local.ColorModel:
		return errors.New(fmt.Sprintf("Unmatched colorModel! Expect : %s, But get %s\n",
			local.ColorModel, server.ColorModel))
	case server.Format != local.Format:
		return errors.New(fmt.Sprintln("Unmatched format! Expect : %s, But get %s\n",
			local.Format, server.Format))
	default:
		return nil
	}
	return nil
}

func extractEfh(url string) string {
	idx := strings.LastIndex(url, "/")
	return string([]byte(url)[idx+1:])
}

func (self *Fileop) cookUrl(url string) string {
	efh := extractEfh(url)
	fhInfo := decodeFh(efh)
	fsize := strconv.FormatInt(fhInfo.Fsize, 10)
	base64Fh := base64.URLEncoding.EncodeToString(fhInfo.Fhandle)
	reqUrl := self.Fopd + "/op?fh=" + base64Fh + "&fsize=" + fsize + "&cmd="
	return reqUrl
}

func (self *Fileop) Mon() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgDownloadUrl()
	msg1 = fmt.Sprintln("Fileop --> doTestImgUrl   error : ", err)
	if err != nil {
		self.DoLog(msg1)
		return
	}
	msg += msg1

	url = self.cookUrl(url)
	err = self.doTestGetImgInfo(url)
	msg1 = fmt.Sprintln("Fileop --> doTestImgInfo  error : ", err)
	if err != nil {
		self.DoLog(msg1)
		return
	}
	msg += msg1
	// {"error":"invalid argument"} // to be settled
	/*err = fp.doTestGetImgEXIF(url)
	msg += fmt.Sprintln("Fileop --> doTestImgExif  : ", err)
	if err != nil {
		return
	}*/

	for _, fileopReq := range self.FileopReq {
		err = self.doTestFileopReq(url, fileopReq)
		msg1 = fmt.Sprintln("Fileop --> "+fileopReq.Name+" error : ", err)
		if err != nil {
			self.DoLog(msg1)
			return
		}
		msg += msg1
	}

	return
}
