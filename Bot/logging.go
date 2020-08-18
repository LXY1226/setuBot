package Bot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

var conf LoggingConf
var logF *os.File
var info *log.Logger

func INFO(a ...interface{}) string {
	return info.Output(2, fmt.Sprint(a...)).Error()
}

func INFOf(f string, a ...interface{}) string {
	return info.Output(2, fmt.Sprintf(f, a...)).Error()
}

func ERROR(a ...interface{}) string {
	return log.Output(2, fmt.Sprint(a...)).Error()
}

func ERRORf(f string, a ...interface{}) string {
	return log.Output(2, fmt.Sprintf(f, a...)).Error()
}

type LogWriter byte // dummy

func (w LogWriter) Write(p []byte) (n int, err error) {
	//p2 := bytes.ReplaceAll(p, []byte("\r"), []byte("\\r"))
	//p2 = bytes.ReplaceAll(p2, []byte("\n"), []byte("\\n"))
	_, _ = os.Stdout.Write(p)
	n, err = logF.Write(p)
	if err == nil {
		err = errors.New(string(p[9 : len(p)-1]))
	}
	return
}

func init() {
	Init(Conf.Logging)
	log.SetOutput(LogWriter(0))
	log.SetFlags(log.Ltime | log.Lshortfile)
	info = log.New(LogWriter(0), "", log.Ltime)
}

func Init(cnf LoggingConf) {
	conf = cnf
	if !conf.Enable {

		return
	}
	err := os.Mkdir(conf.Dir, 0644)
	if err != nil && !os.IsExist(err) {
		ErrOrExit("无法创建日志文件夹", err.Error())
	}
	newLog()
}

func newLog() {
	t := time.Now()
	y, m, d := t.Date()
	time.AfterFunc(time.Date(y, m, d+1, 0, 0, 0, 1, time.Local).Sub(t), newLog)
	y = y % 100
	f, err := os.OpenFile(fmt.Sprintf(conf.Dir+"/%02d-%02d-%02d.log", y, m, d), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		ErrOrExit("无法创建日志文件", err.Error())
		return
	}
	logF, f = f, logF
	if f != nil {
		f.Close()
	}
}
