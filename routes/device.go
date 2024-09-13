package routes

import (
	"MYSQL/controllers"
	"github.com/gin-gonic/gin"
)

func DevicesGroup(router *gin.Engine) {
	deviceGroup := router.Group("/device")
	{
		deviceGroup.POST("/", controllers.AddDevice)         //创建新设备
		deviceGroup.PUT("/:id", controllers.UpdateOwner)     //更新设备所有者
		deviceGroup.GET("/", controllers.GetDescribe)        //获取设备描述信息
		deviceGroup.GET("/:page", controllers.GetALLDevices) //获得所有设备名字与id
		deviceGroup.DELETE("/:id", controllers.DeleteDevice) //删除选定id的设备
	}

}
