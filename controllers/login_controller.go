package controllers

import (
	"MYSQL/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

// 定义一个密钥（可以是任意字符串）
var jwtSecret = []byte("secret_key")

// JWT payload 结构
type Claims struct {
	Username           string `json:"username"`
	jwt.StandardClaims        //Claim是有效负载，用来储存Token
}

// 模拟用户数据,待会删
var userDB map[string]string

// 登录处理函数
func Login(c *gin.Context) {
	var loginData map[string]string
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	userDB = make(map[string]string)
	userDB["username"] = loginData["username"]
	username := loginData["username"]
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln(err)
	}
	password := loginData["password"]
	userDB["password"], err = database.QueryPerson(username, db, 1)
	if err != nil {
		log.Fatalln("用户不存在", err)
	}
	storedPassword := userDB[username]
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	// 验证用户名和密码
	if err == nil {
		// 生成 JWT
		expirationTime := time.Now().Add(2 * time.Hour) // 设置 Token 有效期
		claims := &Claims{
			Username: username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		// 返回 JWT Token
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	}
}

type LogData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func SignUp(c *gin.Context) {
	var loginData LogData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		// 如果绑定失败，返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	database.AddPerson(loginData.Username, loginData.Password, db)
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}
