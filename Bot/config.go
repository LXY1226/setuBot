package Bot

import (
	"crypto/md5"
	"encoding/base64"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"sync"
)

var json = jsoniter.ConfigFastest
var once sync.Once

type (
	Config struct {
		Account AccountConf
		Logging LoggingConf
		App     interface{}
		//App MainConf       `json:"app"`
	}
	AccountConf struct {
		QQ       int64
		Password string
		PassStr  string
		PassByte [16]byte `json:"-"`
	}
	LoggingConf struct {
		Enable       bool
		FileLevel    string
		Dir          string
		ContactOwner bool
	}
)

var Conf = Config{
	Account: AccountConf{
		QQ:       929961096,
		Password: "Password",
	},
	Logging: LoggingConf{
		Enable:       true,
		Dir:          "log",
		ContactOwner: true,
	},
}

var Inited = false

func init() {
	load()

	Inited = true
}

func load() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			save()
		}
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		ErrOrExit("错误的配置文件，请修改或删除配置文件并重启以恢复默认", err.Error())
	}
	if Conf.Account.Password != "" {
		hash := md5.Sum([]byte(Conf.Account.Password))
		Conf.Account.PassByte = hash
		Conf.Account.PassStr = base64.RawStdEncoding.EncodeToString(hash[:])
		Conf.Account.Password = ""
	}
	save()
}

func save() {
	_, err := base64.RawStdEncoding.Decode(Conf.Account.PassByte[:], []byte(Conf.Account.PassStr))
	if err != nil {
		ErrOrExit("无法读取MD5，请不要乱改", err.Error())
	}
	data, _ := json.MarshalIndent(&Conf, "", "    ")
	err = ioutil.WriteFile("config.json", data, 0755)
	if err != nil {
		ErrOrExit("无法写入配置文件", err.Error())
	}
}

func ErrOrExit(msg ...string) {
	if !Inited {
		for _, m := range msg {
			print(m)
		}
		os.Exit(1)
	}
	INFO(msg)
}
