package common

/**
 * log模块封装，用于整个框架中任何地方输出log
 * 在需要的位置使用Info,Debug,Warn,Error即可输出log
 * 各级别日志将根据配置灵活输出
 */

import (
	"fmt"
	"time"
)

// logInfo是为了可变参数输出而定义的接口数据类型
type logInfo []interface{}

// 全部日志的基本输出格式定义
func log(msg ...interface{}) {
	fmt.Println(time.Now().Format("2016-01-01 01:01:01.000"), msg[0], msg[1])
}

// Info log
func Info(msg ...interface{}) {
	msg = append(logInfo{"[INFO]"}, msg...)
	log(msg...)
}

// Debug log
func Debug(msg ...interface{}) {
	msg = append(logInfo{"[DEBUG]"}, msg...)
	log(msg...)
}

// Warn log
func Warn(msg ...interface{}) {
	msg = append(logInfo{"[WARN]"}, msg...)
	log(msg...)
}

// Error log
func Error(msg ...interface{}) {
	msg = append(logInfo{"[ERROR]"}, msg...)
	log(msg...)
}
