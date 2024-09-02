package router

import "github.com/gin-gonic/gin"

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 用户模块
	etcd := r.Group("/etcd")
	{
		//etcd.GET("/:id", userHandler.GetUser)
		//etcd.POST("/", userHandler.CreateUser)
	}

	return r
}
