package main

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net"
	"sync/atomic"
	"time"
)

const (
	clientId     = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	clientSecret = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	loginSecret  = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"
)

type IllustInfo struct {
	Illust struct {
		Title string `json:"title"`
		User  struct {
			ID int `json:"id"`
		} `json:"user"`
		Tags []struct {
			Name           string `json:"name"`
			TranslatedName string `json:"translated_name"`
		} `json:"tags"`
		PageCount      int `json:"page_count"`
		MetaSinglePage struct {
			OriginalImageURL string `json:"original_image_url"`
		} `json:"meta_single_page"`
		MetaPages []struct {
			ImageUrls struct {
				SquareMedium string `json:"square_medium"`
				Medium       string `json:"medium"`
				Large        string `json:"large"`
				Original     string `json:"original"`
			} `json:"image_urls"`
		} `json:"meta_pages"`
		//Width int `json:"width"`
		//Height int `json:"height"`
	} `json:"illust"`
}

type pixivUser struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Errors       *struct {
		System struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"system"`
	} `json:"errors,omitempty"`
	header *fasthttp.RequestHeader
}

var pN = new(uint32)
var pS = []string{
	"210.140.131.226:443",
	"210.140.131.219:443",
	"210.140.131.223:443",
}

func getPixivSrv() string {
	atomic.CompareAndSwapUint32(pN, uint32(len(pS)), 0)
	return pS[atomic.AddUint32(pN, 1)]
}

var pixivClient = &fasthttp.Client{
	Name: "PixivAndroidApp/5.0.64 (Android 10.0)",
	TLSConfig: &tls.Config{
		ServerName:         "-",
		InsecureSkipVerify: true,
	},
	Dial: func(addr string) (net.Conn, error) {
		return fasthttp.Dial(getPixivSrv())
	},
}

func LoadOrLogin(username, password string) *pixivUser {
	u := new(pixivUser)
	data, err := ioutil.ReadFile("pixiv.conf")
	if err != nil {
		u = Login(username, password)
		goto init
	}
	err = json.Unmarshal(data, u)
	if err != nil {
		println(err)
		return Login(username, password)
	}
	u.header = new(fasthttp.RequestHeader)
init:
	u.header.Set("Accept-Language", "zh-cn")
	u.header.Set("Authorization", "Bearer "+u.AccessToken)
	return u
}

func Login(username, password string) *pixivUser {
	const grantStr = "get_secure_url=1&grant_type=password&client_id=" + clientId + "&client_secret=" + clientSecret + "&username="
	req := authReq()
	resp := fasthttp.AcquireResponse()
	u := new(pixivUser)

	buf := make([]byte, 200)[:0]
	buf = append(buf, grantStr...)
	buf = append(buf, username...)
	buf = append(buf, "&password="...)
	buf = append(buf, password...)
	req.SetBody(buf)
	err := pixivClient.Do(req, resp)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(resp.Body(), u)
	if err != nil {
		println("Pixiv API error", err.Error())
		goto clean
	}
	if u.Errors != nil {
		println("Pixiv API error", u.Errors.System.Code, u.Errors.System.Message)
		goto clean
	}
	u.Save()
clean:
	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)
	return u
}

func authReq() *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	XClientTime := time.Now().Format("2006-01-02T15:04:05-07:00")
	h := md5.Sum([]byte(XClientTime + loginSecret))

	req.Header.Set("Accept-Language", "zh-cn")
	req.Header.Set("X-Client-Time", XClientTime)
	req.Header.Set("X-Client-Hash", hex.EncodeToString(h[:]))
	req.Header.SetMethod("POST")
	req.SetRequestURI("https://oauth.secure.pixiv.net/auth/token")
	return req
}

func (u *pixivUser) refreshToken() {
	const grantStr = "get_secure_url=1&grant_type=refresh_token&client_id=" + clientId + "&client_secret=" + clientSecret + "&refresh_token="
	req := authReq()
	resp := fasthttp.AcquireResponse()
	buf := make([]byte, 64)[:0]
	buf = append(buf, grantStr...)
	buf = append(buf, u.RefreshToken...)
	req.SetBody(buf)
	err := pixivClient.Do(req, resp)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(resp.Body(), u)
	if err != nil {
		println("Pixiv API error", err.Error())
		goto clean
	}
	if u.Errors != nil {
		println("Pixiv API error", u.Errors.System.Code, u.Errors.System.Message)
		goto clean
	}
	u.Save()
clean:
	fasthttp.ReleaseResponse(resp)
	fasthttp.ReleaseRequest(req)
	u.header.Set("Authorization", "Bearer "+u.AccessToken)
}

func (u *pixivUser) Save() {
	data, _ := json.Marshal(u)
	err := ioutil.WriteFile("pixiv.conf", data, 0644)
	if err != nil {
		println("无法写入Pixiv用户信息", err.Error())
	}
}
