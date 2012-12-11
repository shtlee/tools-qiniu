package mons

import (
    "qbox.me/mons/example"
    "qbox.me/mons/up"
)

type Interface interface {
	Init(conf []byte) error
	Mon() (msg string, err error)
}


var Mons = map[string]func() Interface{
	"example":   func() Interface { return &example.Example{} },
	"mk_blk" :	 func() Interface { return &up.UpMkBlk{}},
	/*"old_mon":   func() Interface { return &Old{} },
	"rs":        func() Interface { return &Rs{} },
	"rs_upload": func() Interface { return &RsUpload{} },
	"publish":   func() Interface { return &Publish{} },
	"shell":     func() Interface { return &Shell{} },
	"up":        func() Interface { return &Up{} },*/
}
