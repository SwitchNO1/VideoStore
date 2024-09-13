package controllers

import (
	"MYSQL/database"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type Data struct {
	Username string `json:"username" binding:"required"`
	Userid   string
}

func GetUsers(c *gin.Context) {
	var save Data
	err := c.ShouldBindJSON(&save)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	save.Userid, err = database.QueryPerson(save.Username, db, 0)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"userid": save.Userid,
		"username": save.Username})
}
func GetALLUsers(c *gin.Context) {
	// 获取路径参数 :page
	pageParam := c.Param("page")
	// 将 page 参数转换为整数
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		// 如果转换失败，返回错误响应
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	var personInfo []database.Name
	personInfo, err = database.QueryAllPerson(page, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"users": personInfo,
	})
}

func UpdateUser(c *gin.Context) {
	name := c.Param("name")
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.ChangePassword(name, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated", "name": name})
}

func DeleteUser(c *gin.Context) {
	name := c.Param("name")
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.DeletePerson(name, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted", "name": name})
}
