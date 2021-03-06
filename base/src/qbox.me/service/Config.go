package service

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Host           map[string]string `json:"HOST"`
	HostIp         map[string]string `json:"HOST_IP"`

	AccessKey      string    `json:"QBOX_ACCESS_KEY"`
	SecretKey      string    `json:"QBOX_SECRET_KEY"`
	
	BlockBits      uint      `json:"BLOCK_BITS"`
	PutChunkSize   int       `json:"RPUT_CHUNK_SIZE"`
	RPutRetryTimes int       `json:"RPUT_RETRY_TIMES"`

	DataPath     string      `json:"DataPath"`

	Client       string      `json:"CLIENT"`
	ClientId     string      `json:"CLIENT_ID"`
	ClientSecret string      `json:"CLIENT_SECRET"`

	RedirectURI           string `json:"REDIRECT_URI"`
	AuthorizationEndPoint string `json:"AUTHORIZATION_ENDPOINT"`
	TokenEndPoint         string `json:"TOKEN_ENDPOINT"`
}

func LoadConfig(filename string) (c *Config) {
	var conf Config

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return
	}
	c = &conf
	return
}

func EncodeURI(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}
