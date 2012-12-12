package example 

import (
	"errors"
	"qbox.me/mons/util"
)

type ExampleConf struct {
	Msg string `json:"msg"`
	Err bool   `json:"err"`
}

type Example struct {
	conf ExampleConf
}

func (p *Example) Init(conf []byte) error {
	err := util.LoadConf(&(p.conf), conf)
	if err != nil {
		return err
	}
	return nil
}

func (p *Example) Mon() (msg string, err error) {
	if p.conf.Err {
		return p.conf.Msg, errors.New("example err")
	}
	return p.conf.Msg, nil
}