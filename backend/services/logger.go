package services

import (
	"fmt"
)

// Logger 日志接口
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// logger 全局日志实例
var logger Logger

// SetLogger 设置日志实例
func SetLogger(log Logger) {
	logger = log
}

// logInfo 输出信息日志
func logInfo(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	} else {
		fmt.Printf("[INFO] "+format+"\n", v...)
	}
}

// logError 输出错误日志
func logError(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	} else {
		fmt.Printf("[ERROR] "+format+"\n", v...)
	}
}