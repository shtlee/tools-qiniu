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
	"qbox.me/usecase"
	//"qbox.me/usecase/util"
	"qbox.me/cc"
//	"time"
	"qbox.me/shell/shutil/filepath"
)

type Config struct {
	MaxProcs int    `json:"max_procs"`
	Include  string `json:"cases"`
	Env      string `json:"env"`
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
	cases map[string]usecase.Interface
	env  string 
}

type CaseInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (p *Visitor) VisitDir(path string, fi os.FileInfo) bool { return true }
func (p *Visitor) VisitFile(path string, fi os.FileInfo) {
	log.Info("loading ", path, "...")
	conf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("load err :", path, err)
		os.Exit(1)
	}

	log.Info("reading ", p.env, "...")
	env, err := ioutil.ReadFile(p.env)
	if err != nil {
		log.Error("load err : ", p.env, err)
		os.Exit(1)
	}

	var info CaseInfo
	err = json.Unmarshal(conf, &info)
	if err != nil {
		log.Error("load err :", path, err)
		os.Exit(1)
	}
	if _, ok := p.cases[info.Name]; ok || info.Name == "" {
		log.Error("name err(nil or duplication) :", path, info.Name, info.Type)
		os.Exit(1)
	}
	if info.Type != "null" {
		fun, ok := usecase.Cases[info.Type]
		if !ok {
			log.Error("no such type :", info.Type, info.Name)
			os.Exit(1)
		}

		caseEntry := fun()
		err = caseEntry.Init(conf, env)
		if err != nil {
			log.Error("init err :", info.Name, info.Type, err)
			os.Exit(1)
		}
		p.cases[info.Name] = caseEntry
		log.Info("loaded", info.Name, info.Type)
	}
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
	// load cases
	cases := make(map[string]usecase.Interface)
	filepath.Walk(conf.Include, &Visitor{cases, conf.Env}, nil)


	check := func() bool {
		msg := fmt.Sprintf("begin check ...\n")
		errCount := 0
		count := 0
		for k, v := range cases {
			count++
			log.Info("begin check", k, "...")
			msg += fmt.Sprintf("[%v/%v]process %v <<<\n", count, len(cases), k)
			info, err := v.Test()
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
		msg += fmt.Sprintf("all cases finish[%v/%v] <<<\n", len(cases)-errCount, len(cases))
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
