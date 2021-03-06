package qbox

import (
	"encoding/json"
	"io/ioutil"
	//"math/rand"
	"os"
	. "qbox.me/api"
	"qbox.me/api/eu"
	"qbox.me/api/image"
	"qbox.me/api/pub"
	"qbox.me/api/rs"
	"qbox.me/api/uc"
	"qbox.me/api/up"
	"qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
	"testing"
)

// global testing variables
var (
	testfile   = "gopher.jpg"
	testbucket = "test_bucket"
	testkey    = "gopher.jpg"
)

var (
	eus  *eu.EUService
	ups  *up.Service
	ucs  *uc.UCService
	pubs *pub.Publish
	rss  *rs.RSService
	ims  *image.Fileop
)

func doTestSetWatermark(t *testing.T) {

}

func doTestGetWatermark(t *testing.T) {

}

func doTestImage(t *testing.T) {
	urls := make([]string, 2)
	urls[0] = "www.google.com"
	urls[1] = "www.baidu.com"
	host := "mydomain.qbox.me"
	code, err := pubs.Image(testbucket, urls, host, 0)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestUnimage(t *testing.T) {
	code, err := pubs.Unimage(testbucket)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestInfo(t *testing.T) {
	bi, code, err := pubs.Info(testbucket)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(bi)
}

func doTestAccessMode(t *testing.T) {
	code, err := pubs.AccessMode(testbucket, 1)
	if code/100 != 2 {
		t.Fatal(err)
	}
	code, err = pubs.AccessMode(testbucket, 0)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestSeparator(t *testing.T) {
	code, err := pubs.Separator(testbucket, "-")
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestStyle(t *testing.T) {
	style := "imageMogr/auto-orient/thumbnail/!120x120r/gravity/center/crop/!120x120/quality/80"
	code, err := pubs.Style(testbucket, "small.jpg", style)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestUnstyle(t *testing.T) {
	code, err := pubs.Unstyle(testbucket, "small.jpg")
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestPut(t *testing.T) {
	entryURI := testbucket + ":" + testkey
	mimeType := "application/json"
	f, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	ret, code, err := rss.Put(entryURI, mimeType, f, fi.Size())
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestGet(t *testing.T) {
	entryURI := testbucket + ":" + testkey
	ret, code, err := rss.Get(entryURI, "", "", 0)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestStat(t *testing.T) {
	entryURI := testbucket + ":" + testkey
	ret, code, err := rss.Stat(entryURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestDelete(t *testing.T) {
	entryURI := testbucket + ":" + testkey
	ret, code, err := rss.Stat(entryURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestMkbucket(t *testing.T) {
	bucketname := testbucket + "1"
	code, err := rss.Mkbucket(bucketname)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

/*func doTestDrop(t *testing.T) {
	bucketname := testbucket + "1"
	code, err := rss.Drop(bucketname)
	if code/100 != 2 {
		t.Fatal(err)
	}
}*/

func doTestMove(t *testing.T) {
	srcURI := testbucket + ":" + testkey
	destURI := srcURI + "1"
	code, err := rss.Move(srcURI, destURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
	code, err = rss.Move(destURI, srcURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestCopy(t *testing.T) {
	srcURI := testbucket + ":" + testkey
	destURI := srcURI + "1"
	code, err := rss.Copy(srcURI, destURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
	code, err = rss.Delete(destURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestPublish(t *testing.T) {
	domain := "mydomain.qboxtest.me"
	code, err := rss.Publish(domain, testbucket)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestUnpublish(t *testing.T) {
	domain := "mydomain.qboxtest.me"
	code, err := rss.Unpublish(domain)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestBatcher(t *testing.T) {

	b := rss.NewBatcher()
	entryURI := testbucket + ":" + testkey

	b.Get(entryURI)
	b.Stat(entryURI)
	b.Copy(entryURI, entryURI+"1")
	b.Move(entryURI+"1", entryURI+"2")
	b.Delete(entryURI + "2")

	ret, code, err := b.Do()
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestAntiLeechMode(t *testing.T) {
	code, err := ucs.AntiLeechMode(testbucket, 1)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestAddAntiLeech(t *testing.T) {
	code, err := ucs.AddAntiLeech(testbucket, 1, "12.34.56.*")
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestCleanCache(t *testing.T) {
	code, err := ucs.CleanCache(testbucket)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestDelAntiLeech(t *testing.T) {
	code, err := ucs.DelAntiLeech(testbucket, 1, "12.34.56.*")
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestUpPut(t *testing.T) {
	entryURI := testbucket + ":" + testkey
	testfile = "/home/wangjinlei/HDF.pdf"
	f, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	meta := "this is my test image"
	customer := "qboxtest" // uptoken may contain customer field
	callbackparams := ""
	code, err := ups.Put(entryURI, "application/json", customer, meta, callbackparams, f, fi.Size(), "", nil, nil)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

// Advance resumable put
/*func doTestRPutTask(t *testing.T) {

	tfsize := int64(10 * 1024 * 1024)
	tf, err := ioutil.TempFile("", "RPutTest")
	if err != nil {
		t.Fatal(err)
	}
	defer tf.Close()
	if err = tf.Truncate(tfsize); err != nil {
		t.Fatal(err)
	}

	pf, err := ioutil.TempFile("", "RPutProg")
	if err != nil {
		t.Fatal(err)
	}
	pfn := pf.Name()
	pf.Close()

	entryURI := testbucket + ":" + testkey + "1"
	mimeType := "application/json"
	blockcnt := int((tfsize + (1 << ups.BlockBits) - 1) >> ups.BlockBits)
	prog := make([]up.BlockputProgress, blockcnt)

	chunkNotify := func(blkIdx int, p *up.BlockputProgress) {
		//	t.Log(blkIdx,p)
		if prog[blkIdx].Ctx == "" && rand.Intn(blkIdx+1) > blkIdx/2 {
			p1 := *p
			prog[blkIdx] = p1
		}
	}

	blockNotify := func(blkIdx int, p *up.BlockputProgress) {
		//	t.Log(blkIdx,p)
	}

	t1 := ups.NewRPtask(entryURI, mimeType, "", "", "", tf, tfsize)

	for i := 0; i < blockcnt/2; i++ {
		t1.PutBlock(i)
	}

	if err = up.SaveProgress(&prog, pfn); err != nil {
		t.Fatal(err)
	}

	code, err := t1.Run(10, 10, pfn, chunkNotify, blockNotify)
	if code/100 != 2 {
		t.Fatal(err)
	}

	entry, code, err := rss.Get(entryURI, "", "", 0)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(entry)
	code, err = rss.Delete(entryURI)
	if code/100 != 2 {
		t.Fatal(err)
	}
}
*/
func doTestImageInfo(t *testing.T) {
	url1 := "http://qiniuphotos.dn.qbox.me/gogopher.jpg"
	ret, code, err := ims.Info(url1)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestImageExif(t *testing.T) {
	url1 := "http://qiniuphotos.dn.qbox.me/gogopher.jpg"
	ret, code, err := ims.Exif(url1)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(ret)
}

func doTestImageView(t *testing.T) {
	url1 := "http://qiniuphotos.dn.qbox.me/gogopher.jpg"
	f, err := ioutil.TempFile("", "imageview.jpg")
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Name()
	defer f.Close()
	defer os.Remove(fn)

	p := map[string]string{
		"Mode":   "1",
		"Width":  "200",
		"Height": "200",
		"Format": "gif",
		// "Sharpen": "",
		// "HasWatermark": "",
	}
	code, err := ims.View(f, url1, p)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestImageMogr(t *testing.T) {
	f, err := ioutil.TempFile("", "imagemogr")
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Name()
	defer f.Close()
	defer os.Remove(fn)

	entryURI := testbucket + ":" + testkey
	ret, code, err := rss.Get(entryURI, "", "", 0)
	if code/100 != 2 {
		t.Fatal(err)
	}

	p := map[string]string{
		"Thumbnail": "!100x100",
		"Gravity":   "center",
		"Crop":      "!100x100",
		"Quality":   "80",
	}
	code, err = ims.Mogr(f, ret.URL, p)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestImageMogrSaveAs(t *testing.T) {
	var (
		hash image.ImageHash
	)
	f, err := ioutil.TempFile("", "imagemogr")
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Name()
	defer f.Close()
	defer os.Remove(fn)

	entryURI := testbucket + ":" + testkey
	ret, code, err := rss.Get(entryURI, "", "", 0)
	if code/100 != 2 {
		t.Fatal(err)
	}

	p := map[string]string{
		"Thumbnail": "!100x100",
		"Gravity":   "center",
		"Crop":      "!100x100",
		"Quality":   "80",
		"SaveAs":    entryURI,
	}
	code, err = ims.Mogr(&hash, ret.URL, p)
	if code/100 != 2 {
		t.Fatal(err)
	}
	t.Log(hash)
}

func doTestInit(t *testing.T) {
	var (
		c Config
	)

	// load config
	b, err := ioutil.ReadFile("qbox.conf")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}

	// create auth transport layer, uptoken and digest
	auth := &uptoken.AuthPolicy{}
	token := uptoken.MakeAuthTokenString(c.AccessKey, c.SecretKey, auth)
	// uptoken transport
	ut := uptoken.NewTransport(token, nil)

	// digest transport
	dt := digest.NewTransport(c.AccessKey, c.SecretKey, nil)

	if eus, err = eu.New(&c, dt); err != nil {
		t.Fatal(err)
	}
	if pubs, err = pub.New(&c, dt); err != nil {
		t.Fatal(err)
	}
	if rss, err = rs.New(&c, dt); err != nil {
		t.Fatal(err)
	}
	if ucs, err = uc.New(&c, dt); err != nil {
		t.Fatal(err)
	}
	if ups, err = up.New(&c, ut); err != nil {
		t.Fatal(err)
	}
	if ims, err = image.New(&c, dt); err != nil {
		t.Fatal(err)
	}
}

func TestDo(t *testing.T) {

	doTestInit(t)

	doTestSetWatermark(t)
	doTestGetWatermark(t)
	doTestImage(t)
	doTestUnimage(t)
	doTestInfo(t)
	doTestAccessMode(t)
	doTestSeparator(t)
	doTestStyle(t)
	doTestUnstyle(t)
	doTestPut(t)
	defer doTestDelete(t)
	doTestGet(t)
	doTestStat(t)
	doTestMove(t)
	//	doTestCopy(t)
	doTestMkbucket(t)
	//doTestDrop(t)
	doTestPublish(t)
	doTestUnpublish(t)
	doTestBatcher(t)

	//	doTestAntiLeechMode(t)  // not suport digest
	//	doTestAddAntiLeech(t)
	//	doTestDelAntiLeech(t)
	//	doTestCleanCache(t)
	doTestUpPut(t)

	//doTestRPutTask(t)

	doTestImageInfo(t)
	doTestImageExif(t)
	doTestImageView(t)
	doTestImageMogr(t)
	doTestImageMogrSaveAs(t)

}
