package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

type Person struct {
	UserId    int      `db:"user_id"`
	Username  string   `db:"username"`
	Authority []Author `db:`
}
type Author struct {
	DevAuthorId int `db:"id""`
}

type Driver struct {
	DriverId  int      `db:"driver_id"`
	MemAdress []MenAdr `db:"menmory_address""`
}
type MenAdr struct {
	MenId   int
	address string
}

func main() {
	db, err := sqlx.Connect("mysql", "zxy:MYSQL@2024@tcp(test1:3306)/zxy")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	createTab := `CREATE TABLE IF NOT EXISTS person(
					UserId INT UNSIGNED PRIMARY KEY,
					Username VARCHAR(255)
					)`
	_, err = db.Exec(createTab)
	if err != nil {
		log.Fatalln(err)
	}

	createAuth := `CREATE TABLE IF NOT EXISTS author(
					DevAuthorId INT UNSIGNED  UNIQUE ,
					personId INT UNSIGNED,
					FOREIGN KEY(personId) REFERENCES person(UserId)
					)`
	_, err = db.Exec(createAuth)
	if err != nil {
		log.Fatalln(err)
	}
	db.Close()
}

//package main
//
//import (
//	"database/sql"
//	"fmt"
//	_ "github.com/go-sql-driver/mysql"
//)
//
//func main() {
//	dsn := "zxy:MYSQL@2024@tcp(test1:3306)/zxy" // 注意这里的 mysql-container 是 MySQL 容器的名称
//	db, err := sql.Open("mysql", dsn)
//	if err != nil {
//		fmt.Println("Error opening database connection:", err)
//		return
//	}
//	defer db.Close()
//
//	err = db.Ping()
//	if err != nil {
//		fmt.Println("Error connecting to the database:", err)
//	} else {
//		fmt.Println("Successfully connected to the database!")
//	}
//}
