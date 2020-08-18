package main

import (
	"bytes"
	"fmt"
	"github.com/LXY1226/setu/Bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var bClient = client.NewClientMd5(Bot.Conf.Account.QQ, Bot.Conf.Account.PassMD5)
var json = jsoniter.ConfigFastest

type loliconApiResp struct {
	Code        int        `json:"code"`
	Msg         string     `json:"msg"`
	Quota       int        `json:"quota"`
	QuotaMinTTL int        `json:"quota_min_ttl"`
	Count       int        `json:"count"`
	Data        []setuJson `json:"data"`
}

var setus *setuList
var setuChan = make(chan *message.SendingMessage, 4)

func main() {
	rsp, err := bClient.Login()
	//os.Exit(0)
	if err != nil {
		panic(err)
	}
	for {
		if !rsp.Success {
			switch rsp.Error {
			case client.NeedCaptcha:
				ioutil.WriteFile("captcha.jpg", rsp.CaptchaImage, 0644)
				Bot.INFO("请查看验证码并输入")
				var code string
				fmt.Scan(&code)
				rsp, err = bClient.SubmitCaptcha(code, rsp.CaptchaSign)
				continue
			case client.UnsafeDeviceError:
				Bot.INFO("账号已开启设备锁，请前往 ", rsp.VerifyUrl)
				return
			case client.OtherLoginError, client.UnknownLoginError:
				Bot.INFO("登录失败: ", rsp.ErrorMessage)
			}
		}
		break
	}
	fetchSETU()
	Bot.INFO("登录成功 欢迎使用:", bClient.Nickname)
	bClient.ReloadFriendList()
	bClient.ReloadGroupList()
	Bot.INFO("共 ", len(bClient.FriendList), " 好友 ", len(bClient.GroupList), " 群")
	Bot.INFO("アトリは、高性能ですから!")
	bClient.OnGroupMessage(RouteMsg)
	bClient.OnPrivateMessageF(func(m *message.PrivateMessage) bool {
		if m.Sender.Uin == 767763591 {
			return true
		}
		return false
	}, RouteOwner)
	initSETU()
	<-make(chan bool)
}

func WrapFunc() {

}

func RouteOwner(c *client.QQClient, m *message.PrivateMessage) {
	//om := message.NewSendingMessage()
	//for _, e := range m.Elements {
	//	switch e.(type) {
	//	case *message.ImageElement:
	//		t := e.(*message.ImageElement)
	//		resp, err := http.Get(t.Url)
	//		if err != nil {
	//			om.Append(message.NewText(Bot.ERRORf("无法下载图片 %v", err)))
	//			goto send
	//		}
	//		img, _, err := image.Decode(resp.Body)
	//		if err != nil {
	//			om.Append(message.NewText(Bot.ERRORf("无法解码图片 %v", err)))
	//			goto send
	//		}
	//		sImg := image.NewRGBA(image.Rect(0, 0, 240, 240))
	//		draw.NearestNeighbor.Scale(sImg, sImg.Rect, img, img.Bounds(), draw.Src, nil)
	//		hImg
	//		w := new(bytes.Buffer)
	//		if err != nil {
	//			panic(err)
	//		}
	//	}
	//}
	//send:
	//	c.SendPrivateMessage(m.Sender.Uin, om)
}

var setulaiJpgMD5 = []byte{0xb4, 0x07, 0xf7, 0x08, 0xa2, 0xc6, 0xa5, 0x06, 0x34, 0x20, 0x98, 0xdf, 0x7c, 0xac, 0x4a, 0x57}
var setuLock = true

func RouteMsg(c *client.QQClient, msg *message.GroupMessage) {
	s := msg.ToString()
	Bot.INFOf("%s[%d]:%s[%d] %s", msg.GroupName, msg.GroupCode, msg.Sender.Nickname, msg.Sender.Uin, s)
	for _, m := range msg.Elements {
		switch m.(type) {
		case *message.ImageElement:
			e := m.(*message.ImageElement)
			if bytes.Equal(e.Md5, setulaiJpgMD5) {

				SETU(c, msg)
			}
		case *message.TextElement:
			s := m.(*message.TextElement).Content
			switch s {
			case "vtest":
				voice(c, msg)
			case "stat":
				mem := new(runtime.MemStats)
				runtime.ReadMemStats(mem)
				c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(
					message.NewText(Bot.INFOf("共有%d张,%d个Tag\n内存：%dKB\n%d个goroutine\n%s", setus.Len(), len(setus.tagArr),
						mem.Alloc/1024,
						runtime.NumGoroutine(),
						lastMsg))))
			case "!ghs":
				//c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText("无色图")))
				SETU(c, msg)
			case "!ping":
				c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText("pong!")))
				//case "reportTime":
				//	reportTime(c, msg)
			}
			if s[:2] == "p?" {
				c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText("无tag")))
				//SendSETU(c, msg, setus.TagRand(s[2:]))
			}
			if msg.Sender.Uin == 767763591 {
				if s[:2] == "p+" {
					AddSETU(c, msg, s[2:])
				}
				if s[:2] == "BV" {
					c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText(BV2av(s[2:]))))
				}
				//if s[:2] == "p*" {
				//	ShowSETU(c, msg, s[2:])
				//}
				if s == "色图开" {
					setuLock = false
				}
				if s == "色图关" {
					setuLock = true
				}
			}
		}
	}
	return
}

const alphabet = "*************************************************" +
	"\r\x0c.\x1f+\x12(\x1c\x05*******6\x14\x0f\x08'9-$*&3*14*5" +
	"\x07\x04\t2\n,\"\x06\x19\x01******\x1a\0358\x03\x18\x00/\x1b" +
	"\x16)\x10*\x0b%\x02#\x15\x11!\0360\0277 \x0e\x13"

func BV2av(str string) string {
	if len(str) != 10 || str[0] != '1' || str[3] != '4' || str[5] != '1' || str[7] != '7' {
		return "啥啊"
	}
	return "av" + strconv.FormatInt(BV2avInt(
		alphabet[str[9]],
		alphabet[str[8]],
		alphabet[str[1]],
		alphabet[str[6]],
		alphabet[str[2]],
		alphabet[str[4]]), 10)
}

func BV2avInt(a, b, c, d, e, f byte) int64 {
	return (int64(a) +
		int64(b)*58 +
		int64(c)*3364 +
		int64(d)*195112 +
		int64(e)*11316496 +
		int64(f)*656356768 - 0x2084007c0) ^ 0x0a93b324
}

func voice(c *client.QQClient, msg *message.GroupMessage) {
	m := message.NewSendingMessage()
	data, _ := ioutil.ReadFile("C:\\Users\\Lin\\Desktop\\t.amr")
	t, err := c.UploadGroupPtt(msg.GroupCode, data, 30)
	if err != nil {
		Bot.INFO("语音上传失败", err)
		return
	}
	m.Ptt = &t.Ptt
	c.SendGroupMessage(msg.GroupCode, m)
}

func SendSETU(c *client.QQClient, msg *message.GroupMessage, sb *setuBinary) {
	c.SendGroupMessage(msg.GroupCode, AppendSETU(c, msg, nil, sb))
}

func AppendSETU(c *client.QQClient, msg *message.GroupMessage, sm *message.SendingMessage, sb *setuBinary) *message.SendingMessage {
	if sm == nil {
		sm = message.NewSendingMessage()
	}
	var img *message.GroupImageElement
	s := sb.Path()
	sm.Append(message.NewText(s + sb.title))
	fname := "setu/" + s
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}()
		os.MkdirAll(fname[:strings.LastIndexByte(fname, '/')], 0644)
		req.Header.Set("Referer", "https://pixiv.net")
		req.SetRequestURI("https://i.pximg.net/img-original/img/" + s)
		err := fasthttp.Do(req, resp)
		if err != nil {
			sm.Append(message.NewText(Bot.ERROR(fname, "访问pixiv图库失败", err)))
			return sm
		}
		if resp.StatusCode() > 300 {
			sm.Append(message.NewText(Bot.ERROR(fname, "涩图请求失败:", resp.StatusCode())))
			return sm
		}
		data = resp.Body()
		err = ioutil.WriteFile(fname, data, 0755)
		if err != nil {
			Bot.ERROR("不能写入涩图", fname, err)
		}
	}
	img, err = c.UploadGroupImage(msg.GroupCode, data)
	if err != nil {
		sm.Append(message.NewText(Bot.ERROR("上传失败 ", err)))
		return sm
	}
	sm.Append(img)
	return sm
}

func AddSETU(c *client.QQClient, msg *message.GroupMessage, str string) {
	info := new(IllustInfo)
	s := new(setuJson)
	m := message.NewSendingMessage()
	var err error
	var sb *setuBinary
	//var data []byte
	p := 0
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	pixiv.header.CopyTo(&req.Header)
	if i := strings.LastIndexByte(str, '_'); i != -1 {
		p, err = strconv.Atoi(str[i+2:])
		if err != nil {
			m.Append(message.NewText(Bot.ERROR("解析p出错 ", err)))
			goto send
		}
		req.SetRequestURI("https://app-api.pixiv.net/v1/illust/detail?illust_id=" + str[:i])
	} else {
		req.SetRequestURI("https://app-api.pixiv.net/v1/illust/detail?illust_id=" + str)
	}
	err = pixivClient.Do(req, resp)
	if err != nil {
		m.Append(message.NewText(Bot.ERROR("访问pixiv出错 ", err)))
		goto send
	}
	if resp.StatusCode() == 400 {
		pixiv.refreshToken()
		AddSETU(c, msg, str)
		goto clean
	}
	err = json.Unmarshal(resp.Body(), info)
	if err != nil {
		m.Append(message.NewText(Bot.ERROR("解析Pixiv JSON出错 ", err)))
		goto send
	}
	s.Title = info.Illust.Title
	s.Tags = []string{}
	s.UID = uint32(info.Illust.User.ID)
	for _, t := range info.Illust.Tags {
		s.Tags = append(s.Tags, t.Name)
		if t.TranslatedName != "" {
			s.Tags = append(s.Tags, t.TranslatedName)
		}
	}
	if info.Illust.PageCount == 1 {
		s.URL = info.Illust.MetaSinglePage.OriginalImageURL
	} else {
		s.URL = info.Illust.MetaPages[p].ImageUrls.Original
	}
	sb = setus.Transform(s)
	if setus.Add(sb) {
		m.Append(message.NewText(Bot.INFO("添加成功 ", s.Title, s.Path, "\n")))
	} else {
		m.Append(message.NewText(Bot.INFO("已存在 ", s.Title, s.Path, "\n")))
	}
	AppendSETU(c, msg, m, sb)
send:
	c.SendGroupMessage(msg.GroupCode, m)
clean:
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

//var banGroup = map[int64]bool{672076603: true}

func SETU(c *client.QQClient, msg *message.GroupMessage) {
	if setuLock {
		c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText("色图没")))
		return
	}
	//if i, ok := banGroup[msg.GroupCode]; ok && i {
	//	m := message.NewSendingMessage()
	//	m.Append(message.NewText("色图!? 哪呢哪呢"))
	//	c.SendGroupMessage(msg.GroupCode, m)
	//	return
	//}

	m := <-setuChan
	c.SendGroupMessage(msg.GroupCode, m)
}

var apiReq = new(fasthttp.Request)

func initSETU() {
	_ = os.Mkdir("setu", 0644)
	apiReq.SetRequestURI(API)
	setus = NewList()
	go loopSETU()
}

func loopSETU() {
	req := new(fasthttp.Request)
	req.Header.Set("Referer", "https://pixiv.net")
	resp := new(fasthttp.Response)
	for {
		m := message.NewSendingMessage()
		s := setus.Rand().Path()
		fname := "setu/" + s
		data, err := ioutil.ReadFile(fname)
		m.Append(message.NewText(fname[strings.LastIndexByte(fname, '/')+1:]))
		if err != nil {
			os.MkdirAll(fname[:strings.LastIndexByte(fname, '/')], 0644)
			req.SetRequestURI("https://i.pximg.net/img-original/img/" + s)
			err := fasthttp.Do(req, resp)
			if err != nil {
				Bot.ERROR(fname, "访问pixiv图库失败", err)
				continue
			}
			if resp.StatusCode() > 300 {
				Bot.ERROR(fname, "涩图请求失败:", resp.StatusCode())
				continue
			}
			data = resp.Body()
			err = ioutil.WriteFile(fname, data, 0755)
			if err != nil {
				Bot.ERROR("不能写入涩图", fname, err)
			}
		}
		img, err := bClient.UploadGroupImage(624986638, data)
		if err != nil {
			Bot.ERROR("弹药填装失败 ", err)
		} else {
			resp.ResetBody()
			m.Append(img)
			setuChan <- m
			Bot.ERROR("弹药填装完成")
		}
	}
}

var lastMsg string

func fetchSETU() {
	var apiResp loliconApiResp
	resp := fasthttp.AcquireResponse()
	err := fasthttp.Do(apiReq, resp)
	delay := 1 * time.Minute
	if err != nil {
		lastMsg = Bot.INFO("无法调用api", err)
		goto next
	}
	if resp.StatusCode() == 429 {
		lastMsg = Bot.INFO("达到调用额度限制")
		delay = 5 * time.Minute
		goto next
	}
	err = json.Unmarshal(resp.Body(), &apiResp)
	if err != nil {
		lastMsg = Bot.INFO("无法解析json", err.Error())
		println(string(resp.Body()))
		delay = 5 * time.Minute
		goto next
	}
	switch apiResp.Code {
	case 0:
		i := 0
		for _, s := range apiResp.Data {
			sb := setus.Transform(&s)
			if setus.Add(sb) {
				i++
			}
		}
		lastMsg = Bot.INFO("搞到 ", i, " 张涩图")
	case 401:
		lastMsg = Bot.INFO("APIKEY 不存在或被封禁", apiResp.Msg)
		return
	case 403:
		lastMsg = Bot.INFO("由于不规范的操作而被拒绝调用", apiResp.Msg)
		return
	case 404:
		lastMsg = Bot.INFO("找不到符合关键字的色图 [?]", apiResp.Msg)
	case 429:
		delay = time.Duration(apiResp.QuotaMinTTL)*time.Second + time.Minute
		lastMsg = Bot.INFO("达到调用额度限制，下次请求:", delay)
	}
next:
	time.AfterFunc(delay, fetchSETU)
}
