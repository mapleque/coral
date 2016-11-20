package log

/**
 * log模块封装，用于整个框架中任何地方输出log
 * 在需要的位置使用Info,Debug,Warn,Error即可输出log
 * 各级别日志将根据配置灵活输出
 */

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type LEVEL int32

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

type LogPool struct {
	Pool map[string]*Logger
}

type Logger struct {
	path      string
	filename  string
	logFile   *os.File
	maxLevel  LEVEL
	minLevel  LEVEL
	maxNumber int32
	maxSize   int64

	suffix int32
	mux    *sync.RWMutex
	logger *log.Logger
}

var Log *LogPool

func InitLog() *LogPool {
	Log = &LogPool{}
	Log.Pool = make(map[string]*Logger)
	return Log
}

func (lp *LogPool) AddLogger(
	name, path string,
	maxNumber int32, maxSize int64,
	maxLevel LEVEL, minLevel LEVEL) {
	logger := &Logger{}

	logger.path = path
	logger.filename = name
	logger.logFile = openFile(path, name)
	logger.maxNumber = maxNumber
	logger.maxSize = maxSize
	logger.maxLevel = maxLevel
	logger.minLevel = minLevel

	logger.suffix = 0
	logger.mux = new(sync.RWMutex)
	logger.logger = log.New(logger.logFile, "", log.Ldate|log.Ltime)

	lp.Pool[name] = logger
}

// logInfo是为了可变参数输出而定义的接口数据类型
type logInfo []interface{}

func baseLog(level LEVEL, prefix string, msg ...interface{}) {
	msg = append(logInfo{prefix}, msg...)
	if Log != nil {
		Log.log(level, msg...)
	}
	log.Println(msg...)
}

func Debug(msg ...interface{}) {
	baseLog(DEBUG, "[DEBUG]", msg...)
}

func Info(msg ...interface{}) {
	baseLog(INFO, "[INFO]", msg...)
}

func Warn(msg ...interface{}) {
	baseLog(WARN, "[WARN]", msg...)
	Callstack()
}

func Error(msg ...interface{}) {
	baseLog(ERROR, "[ERROR]", msg...)
	Callstack()
}

func Fatal(msg ...interface{}) {
	baseLog(FATAL, "[FATAL]", msg...)
	Callstack()
	os.Exit(1)
}

func Callstack() {
	msg := getCallstack()
	baseLog(ALL, "", msg...)
}

func (lp *LogPool) log(level LEVEL, msg ...interface{}) {
	for _, logger := range lp.Pool {
		if logger.logger != nil &&
			((logger.maxLevel >= level && logger.minLevel <= level) ||
				level == ALL) {
			logger.rotate()
			logger.mux.RLock()
			defer logger.mux.RUnlock()
			logger.logger.Println(msg...)
		}
	}
}

func (lg *Logger) rotate() {
	curFilename := lg.path + "/" + lg.filename
	if fileSize(curFilename) > lg.maxSize {
		lg.mux.Lock()
		defer lg.mux.Unlock()
		lg.suffix = int32((lg.suffix + 1) % lg.maxNumber)
		if lg.logFile != nil {
			lg.logFile.Close()
		}
		tarFilename := curFilename + "." + strconv.Itoa(int(lg.suffix))
		//is file exist, remove it
		if fileIsExist(tarFilename) {
			os.Remove(tarFilename)
		}
		os.Rename(curFilename, tarFilename)
		lg.logFile = openFile(lg.path, lg.filename)
		lg.logger = log.New(lg.logFile, "", log.Ldate|log.Ltime)
	}
}

func fileIsExist(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}

func fileSize(file string) int64 {
	fileInfo, err := os.Stat(file)
	if err != nil {
		Fatal("log path config error", err.Error())
	}
	return fileInfo.Size()
}

func openFile(path, filename string) *os.File {
	pathInfo, err := os.Stat(path)
	if err != nil {
		Fatal("log path config error", err.Error())
	}
	if !pathInfo.IsDir() {
		Fatal("log path [" + path + "] is not a dir")
	}
	logFile, err := os.OpenFile(
		path+"/"+filename,
		os.O_RDWR|os.O_APPEND|os.O_CREATE,
		0666)
	if err != nil {
		Fatal("open log file error", err.Error())
	}
	return logFile
}

// callstack
func getCallstack() []interface{} {
	var callstack []interface{}
	for skip := 0; ; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		callstack = append(callstack, file+":"+strconv.Itoa(line)+"\n")
	}
	return callstack[:len(callstack)-2]
}
