package controllers

import (
	"MYSQL/database"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type DeviceId struct {
	DeviceId string `json:"device_id" `
	OwnerId  string `json:"owner_id " binding:"required"`
}

func AddDevice(c *gin.Context) {
	type adddevice struct {
		Name     string `json:"name" binding:"required"`
		Describe string `json:"describe" `
		OwnerId  string `json:"owner_id " binding:"required"`
	}
	var Id adddevice
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.AddDevice(Id.Name, Id.Describe, Id.OwnerId, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

func GetDescribe(c *gin.Context) {
	var Id DeviceId
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	describes, err := database.QuerydeviceDescribe(Id.DeviceId, Id.OwnerId, db, 1)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"describe": describes})
}
func GetALLDevices(c *gin.Context) {
	// 获取路径参数 :page
	pageParam := c.Param("page")
	// 将 page 参数转换为整数
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		// 如果转换失败，返回错误响应
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	var Id DeviceId
	err = c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	var personInfo []database.Name
	personInfo, err = database.Querydevice(Id.OwnerId, db, page)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"devices": personInfo,
	})
}

// 原主人id放入URL中，设备id和新所有者id放入结构体中
func UpdateOwner(c *gin.Context) {
	oldownerid := c.Param("id")
	var Id DeviceId
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.ChangeDevOwner(Id.DeviceId, oldownerid, Id.OwnerId, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "device updated", "deviceid": Id.DeviceId})
}

func DeleteDevice(c *gin.Context) {
	deviceid := c.Param("id")
	var Id DeviceId
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.Deletedevice(deviceid, Id.OwnerId, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "device deleted", "deviceid": Id.DeviceId})
}
