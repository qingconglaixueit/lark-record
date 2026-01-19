package main

import (
	"lark-record/handlers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
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

		// 多维表格相关
		api.GET("/bitables", handlers.GetBitables)
		api.GET("/bitables/tables", handlers.GetBitableTables)
		api.GET("/bitables/fields", handlers.GetTableFields)

		// 记录操作
		api.POST("/records", handlers.AddRecord)
		api.GET("/records/check", handlers.CheckRecordStatus)
	}

	// 启动服务器
	log.Println("服务器启动在 :8080 端口")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
