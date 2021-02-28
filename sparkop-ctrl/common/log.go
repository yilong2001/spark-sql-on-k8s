package common

import (
	//"context"
	"bytes"
	"errors"
	//"plugin"
	//"flag"
	"fmt"
	"os"
	//"os/signal"
	//"strings"
	//"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	//"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type LogFormatter struct{}

type logFileWriter struct {
	file     *os.File
	logPath  string
	fileDate string
	appName  string
	encoding string
}

func (p *logFileWriter) Write(data []byte) (n int, err error) {
	if p == nil {
		return 0, errors.New("logFileWriter is nil")
	}
	if p.file == nil {
		return 0, errors.New("file not opened")
	}
 
	//判断是否需要切换日期
	fileDate := time.Now().Format("20060102")
	if p.fileDate != fileDate {
		p.file.Close()
		err = os.MkdirAll(fmt.Sprintf("%s/%s", p.logPath, fileDate), os.ModePerm)
		if err != nil {
			return 0, err
		}
		filename := fmt.Sprintf("%s/%s/%s-%s.log", p.logPath, fileDate, p.appName, fileDate)
 
		p.file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
		if err != nil {
			return 0, err
		}
 
	}
	if p.encoding != "" {
		//dataToEncode := ConvertStringToByte(string(data), p.encoding)
		dataToEncode := []byte(string(data))
		n, e := p.file.Write(dataToEncode)
		return n, e
	}
 
	n, e := p.file.Write(data)
	return n, e
 
}

//格式详情
func (s *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	//timestamp := time.Now().Local().Format("2006-01-02 15:04:05.000")
	timestamp := time.Now().Local().Format("15:04:05.000")
	/*var file string
	var len int
	if entry.Caller != nil {
		file = filepath.Base(entry.Caller.File)
		len = entry.Caller.Line
	}*/
	//fmt.Println(entry.Data)
	//msg := fmt.Sprintf("%s [%s:%d][GOID:%d][%s] %s\n", timestamp, file, len, getGID(), strings.ToUpper(entry.Level.String()), entry.Message)
	msg := fmt.Sprintf("%s %s %s] %s\n", strings.ToUpper(entry.Level.String()), timestamp, findCaller(5), entry.Message)
	return []byte(msg), nil
}
 
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func findCaller(skip int) string {
    file := ""
    line := 0
    for i := 0; i < 10; i++ {
        file, line = getCaller(skip + i)
        if !strings.HasPrefix(file, "logrus") {
            break
        }
    }
    return fmt.Sprintf("%s:%d", file, line)
}
// 这里其实可以获取函数名称的: fnName := runtime.FuncForPC(pc).Name()
// 但是我觉得有 文件名和行号就够定位问题, 因此忽略了caller返回的第一个值:pc
// 在标准库log里面我们可以选择记录文件的全路径或者文件名, 但是在使用过程成并发最合适的,
// 因为文件的全路径往往很长, 而文件名在多个包中往往有重复, 因此这里选择多取一层, 取到文件所在的上层目录那层.
func getCaller(skip int) (string, int) {
    _, file, line, ok := runtime.Caller(skip)
    //fmt.Println(file)
    //fmt.Println(line)
    if !ok {
        return "", 0
    }
    n := 0
    for i := len(file) - 1; i > 0; i-- {
        if file[i] == '/' {
            n++
            if n >= 2 {
                file = file[i+1:]
                break
            }
        }
    }
    return file, line
}

func InitLog(logPath string, appName string, encoding string) {
	fileDate := time.Now().Format("20060102")
	//创建目录
	err := os.MkdirAll(fmt.Sprintf("%s/%s", logPath, fileDate), os.ModePerm)
	if err != nil {
		log.Error(err)
		return
	}
 
	filename := fmt.Sprintf("%s/%s/%s-%s.log", logPath, fileDate, appName, fileDate)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0600)
	if err != nil {
		log.Error(err)
		return
	}
 
	fileWriter := logFileWriter{file, logPath, fileDate, appName, encoding}
	log.SetOutput(&fileWriter)
 
	log.SetReportCaller(true)
	log.SetFormatter(new(LogFormatter))
}

func InitLogDefault() {
	log.SetReportCaller(true)
	log.SetFormatter(new(LogFormatter))
}
