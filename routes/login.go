package routes

import (
	"MYSQL/controllers"
	"github.com/gin-gonic/gin"
)

func Login(router gin.Engine) {
	logGroup := router.Group("/login")
	{
		logGroup.POST("/", controllers.Login)
		logGroup.PUT("/", controllers.SignUp)
	}
}
