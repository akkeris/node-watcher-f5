package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type f5creds struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	LoginProviderName string `json:"loginProviderName"`
}

type F5AuthResponse struct {
	Token struct {
		Token string `json:"token"`
	} `json:"token"`
}

var F5Client *http.Client
var F5url string
var F5token string
var creds f5creds
var f5auth string

func NewToken() {
	fmt.Println("[httpclient] Getting new token")
	payload, err := json.Marshal(creds)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", F5url+"/mgmt/shared/authn/login", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", f5auth)
	req.Header.Add("Content-Type", "application/json")
	resp, err := F5Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var auth F5AuthResponse
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &auth); err != nil {
		panic(err)
	}
	if resp.StatusCode > 299 {
		panic(errors.New(string(body)))
	}
	F5token = auth.Token.Token
}

func Startclient() {
	f5username := os.Getenv("F5_USERNAME")
	f5password := os.Getenv("F5_PASSWORD")
	F5url = os.Getenv("F5_URL")
	F5Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	f5auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(f5username+":"+f5password))
	creds.Username = f5username
	creds.Password = f5password
	creds.LoginProviderName = "tmos"
	NewToken()
}
