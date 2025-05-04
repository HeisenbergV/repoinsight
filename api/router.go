package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(handler *Handler) *gin.Engine {
	router := gin.Default()

	// 添加 Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	v1 := router.Group("/api/v1")
	{
		// 仓库相关路由
		repos := v1.Group("/repositories")
		{
			repos.GET("", handler.GetRepositories)
			repos.GET("/:id", handler.GetRepository)
			repos.GET("/search", handler.SearchRepositories)
		}

		// 系统相关路由
		system := v1.Group("/system")
		{
			system.GET("/status", handler.GetStatus)
		}
	}

	return router
}
