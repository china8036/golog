package golog

import (
	"fmt"
	syslog "log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Recode int = iota //运行记录专用
	Error
	Warning
	Emerg
	Info
	Debug
)

const (
	Max_log_size int64 = 2 * 1024 * 1024 //2M
)

var levelMark = [...]string{
	"Recode",
	"Error",
	"Warning",
	"Emerg",
	"Info",
	"Debug"}

var logs string     //记录
var mux *sync.Mutex //互斥锁
var logLevel int = Debug
var writeIntervalTime time.Duration = 10 //秒为单位 多少秒向文件内写入一次日志

func init() {
	mux = new(sync.Mutex)

	go func() {
		for {
			<-time.NewTimer(time.Second * writeIntervalTime).C
			Write()
		}
	}()
}

//主要实现 log 的io.write 接口
type EvetWriter struct {
}

//实现Write方法
func (e *EvetWriter) Write(p []byte) (n int, err error) {
	Log(string(p), Recode)
	return len(p), nil
}

func GetTrace(skip, total_line int) string {
	var trace string
	for i := skip; i < skip+total_line; i++ { // Skip the expected number of frames
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		trace = trace + fmt.Sprintf("%s:%d\n", file, line)
	}
	return trace
}

func LogError(s string) {

	dinfo := GetTrace(2, 5)
	Log(fmt.Sprintf("%s\n%s", s, dinfo), Error)
}

func LogWarning(s string) {
	Log(s, Warning)
}

func LogEmerg(s string) {
	Log(s, Emerg)
}

func LogInfo(s string) {
	Log(s, Info)
}

func LogDebug(s string) {
	Log(s, Debug)
}

//设置记录级别
func SetLogLevel(level int) {
	if level < 0 {
		return
	}
	logLevel = level
}

//设置记录定时刷入文件时间
func SetWriteIntervalTime(seconds time.Duration) {
	if seconds < 0 {
		return
	}
	writeIntervalTime = seconds
}

//放入内存中
func Log(s string, level int) {
	if level > logLevel {
		return // 没有达到记录级别 直接抛弃
	}
	mux.Lock()
	defer mux.Unlock()
	logs = fmt.Sprintf("%s[%s]\n[%s]%s\n", logs, time.Now().String(), levelMark[level], s)
}

//请求将要关闭时候使用
func Write() {
	if logs == "" {
		return
	}
	lfile, err := GetLogFile()
	if err != nil {
		panic(err.Error())
	}
	defer lfile.Close()
	logrecoder := syslog.New(lfile, "", syslog.LstdFlags|syslog.Lmicroseconds) //每个请求打开关闭一次
	mux.Lock()
	defer mux.Unlock()
	logrecoder.Println(logs)
	logs = "" //清空logs

}

func GetLogFile() (*os.File, error) {
	tim := time.Now()
	err := os.MkdirAll(fmt.Sprintf("data/log/%d/%s/%d", tim.Year(), tim.Month().String(), tim.Day()), os.ModePerm)
	if err != nil {
		return nil, err
	}
	file_pre := fmt.Sprintf("data/log/%d/%s/%d/sys", tim.Year(), tim.Month().String(), tim.Day())
	oldfile := fmt.Sprintf("%s.log", file_pre)
	logfile, err := os.OpenFile(oldfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	file_info, er := os.Stat(oldfile)
	if er == nil {
		if file_info.Size() > Max_log_size {
			os.Rename(oldfile, fmt.Sprintf("%s_%d.log", file_pre, tim.Unix()))
		}
	}

	return logfile, err
}
