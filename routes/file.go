package routes

import (
	"MYSQL/controllers"
	"github.com/gin-gonic/gin"
)

func FileRoutes(router *gin.Engine) {
	// 定义用户路由组
	fileGroup := router.Group("/files")
	{
		fileGroup.GET("/", controllers.GetFilePath)        //获得路径的下载路径
		fileGroup.GET("/:page", controllers.GetALLUsers)   //获得自身用户名和id
		fileGroup.PUT("/:id", controllers.UpdateFileOwner) //修改密码
		fileGroup.DELETE("/:id", controllers.DeleteFile)   //删除用户
		fileGroup.POST("/", controllers.AddFile)           //添加文件
	}
}
