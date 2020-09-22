package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"net/http"
	"os"
)

type f5creds struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	LoginProviderName string `json:"loginProviderName"`
}

var F5Client *http.Client
var F5url string
var F5token string
var creds f5creds
var f5auth string

func Startclient() {
	type f5creds struct {
		Username          string `json:"username"`
		Password          string `json:"password"`
		LoginProviderName string `json:"loginProviderName"`
	}


	f5username := os.Getenv("F5_USERNAME")
	f5password := os.Getenv("F5_PASSWORD")
	F5url = os.Getenv("F5_URL")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	F5Client = &http.Client{Transport: tr}
	data := []byte(f5username + ":" + f5password)
	dstr := base64.StdEncoding.EncodeToString(data)
	f5auth = "Basic " + dstr

	creds.Username = f5username
	creds.Password = f5password
	creds.LoginProviderName = "tmos"
	str, err := json.Marshal(creds)
	if err != nil {
		panic(err)
	}
	jsonStr := []byte(string(str))
	urlStr := F5url + "/mgmt/shared/authn/login"
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonStr))
	req.Header.Add("Authorization", f5auth)
	req.Header.Add("Content-Type", "application/json")
	resp, err := F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyj, _ := simplejson.NewFromReader(resp.Body)
	F5token, _ = bodyj.Get("token").Get("token").String()
}

func NewToken() {
    fmt.Println("Getting new Token")
	str, err := json.Marshal(creds)
	if err != nil {
		panic(err)
	}
	jsonStr := []byte(string(str))
	urlStr := F5url + "/mgmt/shared/authn/login"
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonStr))
	req.Header.Add("Authorization", f5auth)
	req.Header.Add("Content-Type", "application/json")
	resp, err := F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyj, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		panic(err)
	}
	F5token, _ = bodyj.Get("token").Get("token").String()
}
