package Bot

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

/*
import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"math/rand"
	"strconv"
)

type Bot struct {
	client *client.QQClient
	messageFunc []func(c client.QQClient, m []message.IMessageElement) (stop bool)
}

func New(conf AccountConf) *Bot {
	B := new(Bot)
	bot := client.NewClientMd5(conf.QQ, conf.PassMD5)
	_, err := bot.Login()
	if err != nil {

		panic("Not implement")
	}
	I("登录成功 欢迎使用:", bot.Nickname)
	_ = bot.ReloadFriendList()
	_ = bot.ReloadGroupList()
	I("共", strconv.Itoa(len(bot.FriendList)), "好友", strconv.Itoa(len(bot.GroupList)), "群")
	I("アトリは、高性能ですから!")
	bot.OnGroupMessage(bot.routeMsg)
	B.client = bot
	return B
}

func (b Bot)routeMsg(c *client.QQClient, m []message.IMessageElement) {
	for _, f := range b.messageFunc {

		f(c, m)
	}
}

func (b Bot)AddMsgFunc(f func(c *client.QQClient, m []message.IMessageElement) (stop bool)) {

}
*/
