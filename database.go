package database

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
)

func ConnectMysql(params ...string) (*sqlx.DB, error) {
	var dsn string
	if len(params) > 0 {
		dsn = params[0]
	} else {
		dsn = "zxy:MYSQL@2024@tcp(test1:3306)/zxy" // 默认DSN
		fmt.Println("现已使用默认地址")
	}

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}
	return db, nil
}

func AddPerson(Name string, Password string, db *sqlx.DB) {
	name := Name
	password := Password
	uniqueID := uuid.New().String()
	query := `INSERT INTO person(userName,passWord,userId) VALUES (?,?,?)`
	_, err := db.Exec(query, name, password, uniqueID)
	if err != nil {
		log.Fatalln(err)
	}

}

func queryPerson(name string, db *sqlx.DB) (string, error) {
	query := `SELECT userId FROM person where userName=?`
	var userId string
	err := db.Get(&userId, query, name)
	if err != nil {
		println("查询出错")
		return "", err
	}
	return userId, err
}

func DeletePerson(Name string, db *sqlx.DB) {
	userid, err := queryPerson(Name, db)
	if err != nil {
		log.Fatalln(err)
	}
	println("若确定删除用户信息则输入密码：")

}
