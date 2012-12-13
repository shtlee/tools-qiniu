package usecase

import (
    "qbox.me/usecase/example"
    "qbox.me/usecase/up"
    "qbox.me/usecase/fop"
)

type Interface interface {
	Init(conf, env []byte) error
	Test() (msg string, err error)
}


var Cases = map[string]func() Interface{
	"example"             :    func() Interface { return &example.Example{} },
	"resumable_put"       :	   func() Interface { return &up.UpResuPut{} },
	"fop_img_info"        :    func() Interface { return &fop.FopImgInfo{} },
    "fop_img_view"        :    func() Interface { return &fop.FopImgOp{} },
    "fop_img_mogr"        :    func() Interface { return &fop.FopImgOp{} },
	/*"old_mon":   func() Interface { return &Old{} },
	"rs":        func() Interface { return &Rs{} },
	"rs_upload": func() Interface { return &RsUpload{} },
	"publish":   func() Interface { return &Publish{} },
	"shell":     func() Interface { return &Shell{} },
	"up":        func() Interface { return &Up{} },*/
}
