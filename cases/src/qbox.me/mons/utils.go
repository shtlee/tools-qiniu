package  mons 

import (
	"encoding/json"
//	"net/http"
//	"qbox.us/oauth"
)

func LoadConf(c interface{}, data []byte) error {
	err := json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

/*
func GetAuthWithKey(accHost string, user, password string, key string) (t http.RoundTripper, err error) {
	cfg := &oauth.Config{
		ClientId:     key,
		ClientSecret: "<ClientSecret>",
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     accHost + "/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &oauth.Transport{Config: cfg}
	_, _, err = transport.ExchangeByPassword(user, password)
	t = transport
	return
}

func GetAuth(accHost string, user, password string) (t http.RoundTripper, err error) {
	return GetAuthWithKey(accHost, user, password, "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2")
}
*/