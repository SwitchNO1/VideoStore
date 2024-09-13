package database

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"log"
	"strings"
	"syscall"
	"time"
)

// Device
type Device struct {
	DeviceName string `db:"deviceName"`
	Describes  string `db:"describes"`
}

// Name
type Name struct {
	Name string `db:"Name"`
	Id   string `db:"Id"`
}

// Device
type Filename struct {
	FileName string `db:"fileName"`
	FileId   string `db:"fileId"`
}
type Sharing struct {
	shareId  string
	fileId   string
	deviceId string
}

// ConnectMysql
func ConnectMysql(params ...string) (*sqlx.DB, error) {
	var dsn string
	if len(params) > 0 {
		dsn = params[0]
	} else {
		dsn = "zxy:MYSQL@2024@tcp(test1:3306)/zxy" // 默认DSN
		fmt.Println("现已使用默认地址")
	}

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败请输入有效地址: %v", err)
	}
	// 设置连接池参数
	db.SetMaxOpenConns(10)                 // 最大打开的连接数
	db.SetMaxIdleConns(5)                  // 最大空闲连接数
	db.SetConnMaxLifetime(time.Minute * 5) // 每个连接的最大存活时间
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("网络连接失败请检查网络设置: %v", err)
	}
	return db, nil
}

//	Function: AddPerson
//	Description: 用户注册
//
// @param Name: 注册用户
//
//	@param db: 链接数据库对象
//	@return 若正常则为nil
func AddPerson(Name string, Password string, db *sqlx.DB) {
	name := Name
	password, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	uniqueID := uuid.New().String()
	query := `INSERT INTO person(userName,passWord,userId) VALUES (?,?,?)`
	_, err = db.Exec(query, name, password, uniqueID)
	if err != nil {
		log.Println("用户注册失败：", err)
	} else {
		println("用户注册成功")
	}

}

// Function: QueryPerson
// Description: 查询用户
//
// @param name: 登录用户名字
// @param db: 链接数据库对象
// @param sw: 0是用户id，1是密码
// @return 返回查询的项若正常则为nil
func QueryPerson(name string, db *sqlx.DB, sw int) (string, error) { //0是id，1是密码
	query := `SELECT userId,passWord FROM person where userName=?`
	type Personpass struct {
		UserId   string `db:"userId"`
		PassWord string `db:"passWord"`
	}
	var userId Personpass
	err := db.Get(&userId, query, name)
	if err != nil {
		println("查询出错")
		return "", err
	}
	if sw == 0 {
		return userId.UserId, nil
	} else if sw == 1 {
		return userId.PassWord, nil
	} else {
		return "", fmt.Errorf("无效的 sw 参数: %d", sw)
	}

}

func QueryAllPerson(i int, db *sqlx.DB) ([]Name, error) {
	query := `SELECT userName ,userId FROM person ORDER BY userName DESC LIMIT 10 OFFSET ? `
	var device []Name
	offset := (i - 1) * 10
	err := db.Get(&device, query, offset)
	if err != nil {
		println("不存在用户")
		return nil, err
	}
	return device, nil
}

// queryPersonById 通过id找到用户名
//
//	@userID 用户id
//	@db 数据库连接对象
//	@return 没问题就返回nil
func queryPersonById(userID string, db *sqlx.DB) (string, error) { //0是id，1是密码
	query := `SELECT userName FROM person where userId=?`
	var username string
	err := db.Get(&username, query, userID)
	if err != nil {
		println("查询出错")
		return "", err
	}
	return username, nil

}

// GetPassword 让用户输入非回显密码
//
//	@return 没问题就返回nil
func GetPassword() string {
	fmt.Print("请输入密码: ")
	// 使用 term.ReadPassword 从标准输入读取密码，不会回显
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("读取密码时出错:", err)
		return ""
	}
	// 输出换行，避免和用户提示信息在同一行
	fmt.Println()
	// 将读取的密码转换为字符串并移除换行符或其他空白字符
	password := strings.TrimSpace(string(bytePassword))

	return password
}

// DeletePerson 删除用户
//
//	@name 用户名字
//	@db 数据库连接对象
//	@return 没问题就返回nil
func DeletePerson(Name string, db *sqlx.DB) error {
	password, err := QueryPerson(Name, db, 1)
	if err != nil {
		log.Println("查询失败", err)
		return err
	}
	println("若确定删除用户信息则输入密码：")
	inputpassword := GetPassword()
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	_, err = db.Exec("DELETE FROM person WHERE userName = ?", Name)
	if err != nil {
		log.Println("删除用户失败:", err)
		return err
	}
	fmt.Println("用户删除成功，所属的设备文件也删除成功")
	return nil
}

// ChangePassword 修改密码
//
//	@Name 设备名
//	@db 数据库连接对象
//	@return 没错误就返回nil
func ChangePassword(Name string, db *sqlx.DB) error {
	password, err := QueryPerson(Name, db, 1)
	if err != nil {
		log.Println("查询失败,该用户不存在", err)
		return err
	}
	println("请输入原密码：")
	inputpassword := GetPassword()
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	fmt.Println("请输入新密码:")
	fmt.Println()
	newpassword := GetPassword()
	fmt.Println("请再次输入新密码:")
	fmt.Println()
	againpassword := GetPassword()
	if newpassword != againpassword {
		return fmt.Errorf("两次输入的密码不一致")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	_, err = db.Exec("UPDATE person SET passWord = ? WHERE userName = ?", hashedPassword, Name)
	if err != nil {
		log.Println("密码更改失败:", err)
		return err
	}
	fmt.Println("密码更改成功")
	return nil

}

// AddDevice 添加设备
//
//	@Name 设备名
//	@describes 描述信息
//	@owner 所有者名
//	@db 数据库连接对象
//	@return 没错误就返回nil
func AddDevice(Name string, describes string, ownerId string, db *sqlx.DB) error {
	uniqueID := uuid.New().String()
	query := `INSERT INTO devices(deviceName,describes,deviceId,ownerId) VALUES (?,?,?,?)`
	_, err := db.Exec(query, Name, describes, uniqueID, ownerId)
	if err != nil {
		log.Println("设备注册失败：", err)
		return err
	}
	fmt.Println("设备注册成功")
	return nil
}

// Querydevice 查询所有设备
//
//	@ownerId 所有者Id
//	@db 数据库连接
//	@ownerId 所有者Id
//	@return 返回Devicaname结构体，若无错误返回nil
func Querydevice(ownerId string, db *sqlx.DB, page int) ([]Name, error) {
	query := `SELECT deviceName,deviceId FROM devices where ownerId=? ORDER BY deviceName DESC LIMIT 10 OFFSET ? `
	var device []Name
	offset := (page - 1) * 10
	err := db.Get(&device, query, ownerId, offset)
	if err != nil {
		println("不存在该设备")
		return nil, err
	}
	return device, nil
}

// QuerydeviceDescribe 用设备名查询设备id,所有者查询指定的设备
//
//	@deviceId 查询设备id
//	@db 数据库连接
//	@ownerId 所有者Id
//	@sw 0代表设备id，1代表描述信息
//	@return 返回所选数据，若无错误返回nil
func QuerydeviceDescribe(deviceId string, ownerId string, db *sqlx.DB, sw int) (string, error) {
	query := `SELECT deviceName,describes FROM devices where deviceId=? AND ownerId=?`
	var device Device
	err := db.Get(&device, query, deviceId, ownerId)
	if err != nil {
		println("不存在该设备")
		return "", err
	}
	if sw == 0 {
		return device.DeviceName, nil
	} else if sw == 1 {
		return device.Describes, nil
	} else {
		return "", fmt.Errorf("无效的 sw 参数: %d", sw)
	}

}

//	Function: Deletedevice
//	Description: 删除设备
//
// @param deviceId: 设备Id
//
//	@ownerId 所有者Id
//	@param db: 链接数据库对象
//	@return 若正常则为nil
func Deletedevice(deviceId string, ownerId string, db *sqlx.DB) error {
	_, err := db.Exec("DELETE FROM devices WHERE deviceId = ?AND ownerId=?", deviceId, ownerId)
	if err != nil {
		log.Println("删除设备失败:", err)
		return err
	}
	fmt.Println("设备删除成功")
	return nil
}

// Function: ChangeDevOwner
// Description: 改变设备所有者
//
// @param Name: 登录用户
// @param newowner: 新的拥有者
// @param db: 链接数据库对象
// @return 若正常则为nil
func ChangeDevOwner(deviceId string, oldownerId string, newownerId string, db *sqlx.DB) error {
	oldowner, err := queryPersonById(oldownerId, db)
	fmt.Printf("该设备属于用户%s\n", oldowner)
	fmt.Println("若确定修改所有者则输入该用户密码：")
	inputpassword := GetPassword()
	password, err := QueryPerson(oldowner, db, 1)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	_, err = db.Exec("UPDATE devices SET ownerId = ? WHERE deviceId = ? AND ownerId=?", newownerId, deviceId, oldownerId)
	if err != nil {
		log.Println("所有者更新失败:", err)
		return err
	}
	err = updatefileowner(deviceId, oldownerId, newownerId, db)
	if err != nil {
		log.Println("文件所有者更新失败:", err)
		return err
	}
	err = updateshareowner(deviceId, oldownerId, newownerId, db)
	if err != nil {
		log.Println("分享者记录所有者更新失败:", err)
		return err
	}
	fmt.Println("所有者更新成功")
	return nil
}

// Function: updatefileowner
// Description: 改变设备所有者
//
// @param Name: 登录用户
// @param newowner: 新的拥有者
// @param db: 链接数据库对象
// @return 若正常则为nil
func updatefileowner(deviceId string, oldownerid string, newownerid string, db *sqlx.DB) error {
	tx, err := db.Beginx()
	_, err = db.Exec("UPDATE files SET ownerId = ?  WHERE deviceId = ? AND ownerId=?", newownerid, deviceId, oldownerid)
	if err != nil {
		log.Println("所有者更新失败:", err)
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		tx.Rollback()
		return err
	}
	fmt.Println("所有者更新成功")
	return nil
}

// Function: updateshareowner
// Description: 改变分享记录中的所有者
//
// @param deviceId: 设备id
// @param oldownerid: 旧的拥有者id
// @param newownerid: 新的拥有者id
// @param db: 链接数据库对象
// @return 若正常则为nil
func updateshareowner(deviceId string, oldownerid string, newownerid string, db *sqlx.DB) error {
	tx, err := db.Beginx()
	_, err = db.Exec("UPDATE sharing SET sharedOwnerId = ?  WHERE deviceId = ? AND sharedOwnerId=?", newownerid, deviceId, oldownerid)
	if err != nil {
		log.Println("所有者更新失败:", err)
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Println("提交事务失败:", err)
		tx.Rollback()
		return err
	}
	fmt.Println("所有者更新成功")
	return nil
}

// AddFile 添加文件
//
//	@param deviceId 所属设备Id
//	@param fileName 文件名
//	@param filePath 文件的储存路径
//	@param db 数据库连接对象
//	@return 没错误就返回nil
func AddFile(deviceId string, ownerId string, fileName string, filePath string, db *sqlx.DB) error {
	uniqueID := uuid.New().String()
	query := `INSERT INTO files(deviceId,filePath,fileId,ownerId,fileName) VALUES (?,?,?,?,?)`
	_, err := db.Exec(query, deviceId, filePath, uniqueID, ownerId, fileName)
	if err != nil {
		log.Println("储存文件路径失败：", err)
		return err
	}
	fmt.Println("储存文件路径成功")
	return nil
}

// QueryAllFile 查询所有该设备所有文件
//
//	@deviceId 设备ID
//	@ownerId 所有者Id
//	@db 数据库连接
//	@ownerId 所有者Id
//	@return 返回Devicaname结构体，若无错误返回nil
func QueryAllFile(deviceId string, ownerId string, page int, db *sqlx.DB) (*Filename, error) {
	query := `SELECT fileName,fileId FROM devices where ownerId=? AND deviceId=? ORDER BY userName DESC LIMIT 10 OFFSET ? `
	var file Filename
	err := db.Get(&file, query, ownerId, deviceId, page)
	if err != nil {
		println("不存在该设备")
		return &Filename{}, err
	}
	return &file, nil
}

// Function: QueryFile
//
//	Description: 查询
//
//	@param fileId: 查询id
//	@param db: 链接数据库对象
//	@param params:分别返回0:文件名，1:存储路径,请按顺序输入
//	@return string切片存放顺序与数字对应若正常则为nil
func QueryFile(fileId string, db *sqlx.DB, params ...int) ([]string, error) {
	query := `SELECT fileName,filepath FROM files where fileId=?`

	type File struct {
		fileName string `db:"fileName"`
		filepath string `db:"filepath"`
	}
	var file File
	err := db.Get(&file, query, fileId)
	if err != nil {
		println("查询出错或文件不存在")
		return nil, err
	}

	// 根据 params 参数返回相应字段
	var result []string
	for _, param := range params {
		switch param {
		case 0:
			result = append(result, file.fileName) // 返回 fileId
		case 1:
			result = append(result, file.filepath) // 返回 filepath
		default:
			return nil, fmt.Errorf("无效的参数: %d", param)
		}
	}

	return result, nil
}

// Function: DeleteFile
//
//	Description: 批量删除文件
//
//	@param filename: 查询文件名
//	@param db: 链接数据库对象
//	@param filenames:文件名
//	@return 若正常则为nil
func DeleteFile(db *sqlx.DB, fileId ...string) error {
	// 如果没有文件名传入，直接返回
	if len(fileId) == 0 {
		return nil
	}

	// 创建一个 IN 查询条件
	query, args, err := sqlx.In(`DELETE FROM files WHERE fileId IN (?)`, fileId)
	if err != nil {
		fmt.Printf("构建查询失败: %v\n", err)
		return err
	}

	// sqlx.In 会返回一个带有 `?` 的查询，需要 Rebind 以适应数据库方言
	query = db.Rebind(query)
	_, err = db.Exec(query, args...)
	if err != nil {
		fmt.Printf("删除文件失败: %v\n", err)
		return err
	}
	return nil
}

//	 Function: authtenticaton
//	 Description: 各种删除动作前的验证所有者的动作
//
//		@param Name: 登录用户
//	 @param newowner: 新的拥有者
//	 @param db: 链接数据库对象
//	 @return 若正常则为nil
func Authtenticaton(db *sqlx.DB, ownername string) error {

	fmt.Printf("该设备属于用户%s\n", ownername)
	fmt.Println("若确定删除则输入该用户密码：")
	inputpassword := GetPassword()
	password, err := QueryPerson(ownername, db, 1)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	return nil
}

//	 Function: CheckDeviceName
//	 Description: 设备取名前查询是否同名
//
//		@param ownerId: 所有人Id
//	 @param deviceName: 设备名字
//	 @param db: 链接数据库对象
func CheckDeviceName(ownerId string, deviceName string, db *sqlx.DB) {
	query := `SELECT deviceId FROM devices where deviceName=? AND ownerId=? `
	var deviceid string
	err := db.Get(&deviceid, query, deviceName, ownerId)
	if err != nil {
		fmt.Println("该设备名已存在，请修改名字")
	}
}

//	 Function: CheckFileName
//	 Description: 文件取名前查询是否同名
//
//		@param ownerId: 所有人Id
//	 @param fileName: 文件名字
//	 @param db: 链接数据库对象
func CheckFileName(ownerId string, fileName string, db *sqlx.DB) {
	query := `SELECT fileId FROM files where fileName=? AND ownerId=? `
	var fileid string
	err := db.Get(&fileid, query, fileName, ownerId)
	if err != nil {
		fmt.Println("该文件名已存在，请修改名字")
	}
}

// Function: ChangeFileOwn
// Description: 改变分享文件的所有者
//
// @param Name: 登录用户
// @param newowner: 新的拥有者
// @param db: 链接数据库对象
// @return 若正常则为nil
func ChangeFileOwn(fileId string, newdeviceId string, oldownerId string, newownerId string, db *sqlx.DB) error {

	_, err := db.Exec("UPDATE files SET ownerId = ?,deviceId=? WHERE fileId=? AND ownerId=? ", newownerId, newdeviceId, fileId, oldownerId)
	if err != nil {
		log.Println("所有者更新失败:", err)
		return err
	}
	err = updateshareown(fileId, newdeviceId, oldownerId, newownerId, db)
	if err != nil {
		log.Println("分享者记录所有者更新失败:", err)
		return err
	}
	fmt.Println("所有者更新成功")
	return nil
}

func updateshareown(fileId string, newdeviceId string, oldownerId string, newownerId string, db *sqlx.DB) error {
	_, err := db.Exec("UPDATE sharing SET sharedOwnerId = ?,deviceId=? WHERE fileId=? AND sharedOwnerId=? ", newownerId, newdeviceId, fileId, oldownerId)
	if err != nil {
		log.Println("所有者更新失败:", err)
		return err
	}
	fmt.Println("所有者更新成功")
	return nil
}

// Addsharing 分享文件添加
//
//	@param deviceId 所属设备Id
//	@param fileName 文件名
//	@param filePath 文件的储存路径
//	@param db 数据库连接对象
//	@return 没错误就返回nil
func Addsharing(deviceId string, sharedOwnerId string, fileId string, sharedUserId string, db *sqlx.DB) error {
	uniqueID := uuid.New().String()
	query := `INSERT INTO files(deviceId,sharedUserId,fileId,sharedOwnerId,fileId) VALUES (?,?,?,?,?)`
	_, err := db.Exec(query, deviceId, sharedUserId, uniqueID, sharedOwnerId, fileId)
	if err != nil {
		log.Println("储存分享记录：", err)
		return err
	}
	fmt.Println("储存分享记录成功")
	return nil
}

// Function: DeleteFile
//
//	Description: 批量删除文件
//
//	@param filename: 查询文件名
//	@param db: 链接数据库对象
//	@param filenames:文件名
//	@return 若正常则为nil
func DeleteSharingByOwn(db *sqlx.DB, sharedOwnerId string, fileId ...string) error {
	// 如果没有文件名传入，直接返回
	if len(fileId) == 0 {
		return nil
	}

	// 创建一个 IN 查询条件
	query, args, err := sqlx.In(`DELETE FROM sharing WHERE sharedOwnerId=? AND IN (?)`, sharedOwnerId, fileId)
	if err != nil {
		fmt.Printf("构建查询失败: %v\n", err)
		return err
	}

	// sqlx.In 会返回一个带有 `?` 的查询，需要 Rebind 以适应数据库方言
	query = db.Rebind(query)
	_, err = db.Exec(query, args...)
	if err != nil {
		fmt.Printf("删除分享记录失败: %v\n", err)
		return err
	}
	return nil
}
func DeleteSharingByShare(db *sqlx.DB, sharedUserId string, fileId ...string) error {
	// 如果没有文件名传入，直接返回
	if len(fileId) == 0 {
		return nil
	}

	// 创建一个 IN 查询条件
	query, args, err := sqlx.In(`DELETE FROM sharing WHERE sharedUserId=? AND IN (?)`, sharedUserId, fileId)
	if err != nil {
		fmt.Printf("构建查询失败: %v\n", err)
		return err
	}

	// sqlx.In 会返回一个带有 `?` 的查询，需要 Rebind 以适应数据库方言
	query = db.Rebind(query)
	_, err = db.Exec(query, args...)
	if err != nil {
		fmt.Printf("删除分享记录失败: %v\n", err)
		return err
	}
	return nil
}

// QueryAllFile 查询所有该设备所有文件
//
//	@deviceId 设备ID
//	@ownerId 所有者Id
//	@db 数据库连接
//	@ownerId 所有者Id
//	@return 返回Devicaname结构体，若无错误返回nil
func QuerySharing(sharedUserId string, sharedOwnerId string, db *sqlx.DB) (*Sharing, error) {
	query := `SELECT shareId,fileId,deviceId FROM sharing where sharedOwnerId=? AND sharedUserId=?`
	var sharing Sharing
	err := db.Get(&sharing, query, sharedOwnerId, sharedUserId)
	if err != nil {
		println("不存在分享记录")
		return &Sharing{}, err
	}
	return &sharing, nil
}
func QueryAllSharing(sharedUserId string, db *sqlx.DB) (*Sharing, error) {
	query := `SELECT shareId,fileId,deviceId FROM sharing where sharedUserId=?`
	var sharing Sharing
	err := db.Get(&sharing, query, sharedUserId)
	if err != nil {
		println("不存在分享记录")
		return &Sharing{}, err
	}
	return &sharing, nil
}
func CreatePersonTable(db *sqlx.DB) {
	query := "CREATE TABLE IF NOT EXISTS person (" +
		"userId CHAR(36) PRIMARY KEY NOT NULL," +
		"userName VARCHAR(50) UNIQUE NOT NULL," +
		"createAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP," +
		"email VARCHAR(255) NOT NULL," +
		"passWord VARCHAR(255) NOT NULL)"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalln(err)
	}
}

func CreateDeviceTable(db *sqlx.DB) {
	query := "CREATE TABLE IF NOT EXISTS devices (" +
		"deviceId CHAR(36) PRIMARY KEY NOT NULL," +
		"deviceName VARCHAR(50) NOT NULL," +
		"createAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP," +
		"describes VARCHAR(255) DEFAULT 'NULL'," +
		"ownerId CHAR(36) NOT NULL ," +
		"FOREIGN KEY(ownerId) REFERENCES person(userId) ON DELETE CASCADE" +
		")"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalln(err)
	}
}

func CreateFileTable(db *sqlx.DB) {
	query := "CREATE TABLE IF NOT EXISTS files (" +
		"fileId CHAR(36) PRIMARY KEY NOT NULL," +
		"fileName VARCHAR(50)  NOT NULL," +
		"createAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP," +
		"ownerId CHAR(36) NOT NULL," +
		"filePath VARCHAR(255) NOT NULL," +
		"deviceId CHAR(36) NOT NULL ," +
		"FOREIGN KEY(deviceId) REFERENCES devices(deviceId) ON DELETE CASCADE" +
		")"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalln(err)
	}
}
func CreateSharingTable(db *sqlx.DB) {
	query := "CREATE TABLE IF NOT EXISTS sharing (" +
		"shareId CHAR(36) PRIMARY KEY ," +
		"deviceId CHAR(36)  NOT NULL," +
		"createAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP," +
		"sharedOwnerId CHAR(36) NOT NULL," +
		"sharedUserId CHAR(36) NOT NULL," +
		"fileId CHAR(36) NOT NULL ," +
		"FOREIGN KEY(fileId) REFERENCES files(fileId) ON DELETE CASCADE" +
		")"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalln(err)
	}
}
