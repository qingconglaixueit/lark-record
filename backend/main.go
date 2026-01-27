package main

import (
	"fmt"
	"io"
	"lark-record/handlers"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Logger 日志管理器
type Logger struct {
	mu     sync.Mutex
	file   *os.File
	logger *log.Logger
	prefix string
}

// NewLogger 创建新的日志管理器
func NewLogger(prefix string) *Logger {
	logger := &Logger{
		prefix: prefix,
	}
	logger.rotate()
	// 启动定时轮转
	go func() {
		for {
			now := time.Now()
			// 计算到下一天的时间间隔
			nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(nextDay.Sub(now))
			logger.rotate()
		}
	}()
	// 启动日志清理
	go func() {
		for {
			logger.cleanOldLogs()
			// 每天清理一次
			time.Sleep(24 * time.Hour)
		}
	}()
	return logger
}

// rotate 轮转日志文件
func (l *Logger) rotate() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭当前文件
	if l.file != nil {
		l.file.Close()
	}

	// 创建日志目录
	if err := os.MkdirAll(".", 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return
	}

	// 创建新的日志文件
	now := time.Now()
	filename := fmt.Sprintf("%s-%s.log", l.prefix, now.Format("2006-01-02"))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("创建日志文件失败: %v", err)
		return
	}

	l.file = file
	// 设置日志输出到控制台和文件
	l.logger = log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags)
}

// cleanOldLogs 清理旧日志（保留半个月）
func (l *Logger) cleanOldLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取当前时间
	now := time.Now()
	// 计算半个月前的时间
	cutoff := now.AddDate(0, 0, -15)

	// 遍历当前目录
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Printf("读取目录失败: %v", err)
		return
	}

	// 清理旧日志
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// 检查是否是当前前缀的日志文件
		if !strings.HasPrefix(filename, l.prefix+"-") {
			continue
		}

		// 解析日志文件的日期
		if len(filename) < len(l.prefix+"-0000-00-00.log") {
			continue
		}

		dateStr := filename[len(l.prefix)+1 : len(l.prefix)+11]
		logDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// 如果日志文件日期早于截止日期，删除
		if logDate.Before(cutoff) {
			if err := os.Remove(filename); err != nil {
				log.Printf("删除旧日志失败: %v", err)
			} else {
				log.Printf("删除旧日志: %s", filename)
			}
		}
	}
}

// Println 输出日志
func (l *Logger) Println(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logger != nil {
		l.logger.Println(v...)
	}
}

// Printf 格式化输出日志
func (l *Logger) Printf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logger != nil {
		l.logger.Printf(format, v...)
	}
}

// Fatalf 格式化输出日志并退出
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logger != nil {
		l.logger.Fatalf(format, v...)
	}
	os.Exit(1)
}

var logger *Logger

func main() {
	// 初始化日志管理器
	logger = NewLogger("server")

	// 创建Gin路由
	r := gin.Default()

	// 配置CORS，允许Chrome扩展访问
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// 设置路由
	api := r.Group("/api")
	{
		// 配置相关
		api.POST("/config", handlers.SaveConfig)
		api.GET("/config", handlers.GetConfig)
		api.POST("/config/test", handlers.TestConfig)

		// 多维表格相关
		api.GET("/bitables", handlers.GetBitables)
		api.GET("/bitables/tables", handlers.GetBitableTables)
		api.GET("/bitables/fields", handlers.GetTableFields)

		// 记录操作
		api.POST("/records", handlers.AddRecord)
		api.GET("/records/check", handlers.CheckRecordStatus)

		// AI解析
		api.POST("/ai/parse", handlers.AIParse)
		// 获取AI模型列表
		api.GET("/ai/models", handlers.GetAIModels)
	}

	// 启动服务器
	logger.Println("服务器启动在 :8080 端口")
	if err := r.Run(":8080"); err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}