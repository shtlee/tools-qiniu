package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"io/ioutil"
//	"math/rand"
	"encoding/json"
	"qbox.me/log"
	"qbox.me/mons"
	"qbox.me/cc"
//	"time"
	"qbox.me/shell/shutil/filepath"
)

type Config struct {
	MaxProcs int    `json:"max_procs"`
	Include  string `json:"include"`
}

func getConf(file string) (moConf *Config, err error) {
	conf := Config{}
	confData, err := ioutil.ReadFile(file)
	if err == nil {
		err = json.Unmarshal(confData, &conf)
	}
	return &conf, err
}

type Visitor struct {
	mons map[string]mons.Interface
}

type MonInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (p *Visitor) VisitDir(path string, fi os.FileInfo) bool { return true }
func (p *Visitor) VisitFile(path string, fi os.FileInfo) {
	log.Info("loading", path, "...")
	conf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("load err :", path, err)
		os.Exit(1)
	}
	var info MonInfo
	err = json.Unmarshal(conf, &info)
	if err != nil {
		log.Error("load err :", path, err)
		os.Exit(1)
	}
	if _, ok := p.mons[info.Name]; ok || info.Name == "" {
		log.Error("name err(nil or duplication) :", path, info.Name, info.Type)
		os.Exit(1)
	}
	fun, ok := mons.Mons[info.Type]
	if !ok {
		log.Error("no such type :", info.Type, info.Name)
		os.Exit(1)
	}

	mon := fun()
	err = mon.Init(conf)
	if err != nil {
		log.Error("init err :", info.Name, info.Type, err)
		os.Exit(1)
	}
	p.mons[info.Name] = mon
	log.Info("loaded", info.Name, info.Type)
}

func main() {
	confDir, _ := cc.GetConfigDir("qbox.me")
	var confName *string = flag.String("f", confDir+"/qboxtest.conf", "the config file")
	flag.Parse()
	log.Info("Use the config file of " + *confName)
	conf, err := getConf(*confName)
	if err != nil {
		log.Error(err)
		return
	}

	runtime.GOMAXPROCS(conf.MaxProcs)

	//rand.Seed(time.Nanoseconds())
	mons := make(map[string]mons.Interface)
	filepath.Walk(conf.Include, &Visitor{mons}, nil)

	fmt.Println("mons : ", mons)

	check := func() bool {
		msg := fmt.Sprintf("begin check ...\n")
		errCount := 0
		count := 0
		for k, v := range mons {
			count++
			log.Info("begin check", k, "...")
			msg += fmt.Sprintf("[%v/%v]process %v <<<\n", count, len(mons), k)
			info, err := v.Mon()
			log.Info("check done :\n", info, k, err)
			msg += info
			msg += "\n"
			if err != nil {
				msg += fmt.Sprintf("!!!!!!!!!!mon case [%v] err!!![%v]\n", k, err)
				errCount++
			} else {
				msg += fmt.Sprintf("[no err]%v done <<<\n", k)
			}
		}
		msg += fmt.Sprintf("all cases finish[%v/%v] <<<\n", len(mons)-errCount, len(mons))
		fmt.Println("----------------- result ------------------")
		fmt.Println(msg)
		fmt.Println("-------------------------------------------")
		return errCount == 0
	}
	
	b := check()
	if b {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
