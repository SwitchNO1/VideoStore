package routes

import (
	"MYSQL/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	// 定义用户路由组
	userGroup := router.Group("/users")
	{
		userGroup.GET("/", controllers.GetUsers)
		userGroup.GET("/:page", controllers.GetALLUsers)   //获得自身用户名和id
		userGroup.PUT("/:name", controllers.UpdateUser)    //修改密码
		userGroup.DELETE("/:name", controllers.DeleteUser) //删除用户
	}
}
