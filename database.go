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
)

// Device
type Device struct {
	deviceId  string `db:"deviceId"`
	ownerId   string `db:"ownerId"`
	describes string `db:"describes"`
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

// Function: queryPerson
// Description: 查询用户
//
// @param name: 登录用户名字
// @param db: 链接数据库对象
// @param sw: 0是用户id，1是密码
// @return 返回查询的项若正常则为nil
func queryPerson(name string, db *sqlx.DB, sw int) (string, error) { //0是id，1是密码
	query := `SELECT userId,passWord FROM person where userName=?`
	type Person struct {
		UserId   string `db:"userId"`
		PassWord string `db:"passWord"`
	}
	var userId Person
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
	password, err := queryPerson(Name, db, 1)
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
	password, err := queryPerson(Name, db, 1)
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
func AddDevice(Name string, describes string, owner string, db *sqlx.DB) error {
	uniqueID := uuid.New().String()
	ownerId, err := queryPerson(owner, db, 0)
	if err != nil {
		log.Println("不存在该用户：", err)
		return err
	}
	query := `INSERT INTO devices(deviceName,describes,deviceId,ownerId) VALUES (?,?,?,?)`
	_, err = db.Exec(query, Name, describes, uniqueID, ownerId)
	if err != nil {
		log.Println("设备注册失败：", err)
		return err
	}
	fmt.Println("设备注册成功")
	return nil
}

// Querydevice 用设备名查询设备id,所有者id，具体描述
//
//	@name 查询设备名
//	@db 数据库连接
//	@sw 0代表设备id，1代表所有者id，2代表描述信息
//	@return 返回所选数据，若无错误返回nil
func Querydevice(devicename string, owner string, db *sqlx.DB, sw int) (string, error) {
	owId, err := queryPerson(owner, db, 0)
	query := `SELECT deviceId,ownerId,describes FROM devices where deviceName=? AND ownerId=?`
	var device Device
	err = db.Get(&device, query, devicename, owId)
	if err != nil {
		println("不存在该设备")
		return "", err
	}
	if sw == 0 {
		return device.deviceId, nil
	} else if sw == 1 {
		return device.ownerId, nil
	} else if sw == 2 {
		return device.describes, nil
	} else {
		return "", fmt.Errorf("无效的 sw 参数: %d", sw)
	}

}

//	 Function: Deletedevice
//	 Description: 删除设备
//
//		@param Name: 设备名
//	 @param db: 链接数据库对象
//	 @return 若正常则为nil
func Deletedevice(Name string, owner string, db *sqlx.DB) error {
	owId, err := queryPerson(owner, db, 0)
	_, err = db.Exec("DELETE FROM devices WHERE deviceName = ?AND ownerId=?", Name, owId)
	if err != nil {
		log.Println("删除设备失败:", err)
		return err
	}
	fmt.Println("设备删除成功")
	return nil
}

// Function: ChangeOwner
// Description: 改变设备所有者
//
// @param Name: 登录用户
// @param newowner: 新的拥有者
// @param db: 链接数据库对象
// @return 若正常则为nil
func ChangeOwner(Name string, oldowner string, newowner string, db *sqlx.DB) error {
	oldownId, err := queryPerson(oldowner, db, 0)
	fmt.Printf("该设备属于用户%s\n", oldowner)
	fmt.Println("若确定修改所有者则输入该用户密码：")
	inputpassword := GetPassword()
	password, err := queryPerson(oldowner, db, 1)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	newownerId, err := queryPerson(newowner, db, 0)
	if err != nil {
		log.Println("新用户查询失败:", err)
		return err
	}
	deviceid, err := Querydevice(Name, oldowner, db, 0)
	_, err = db.Exec("UPDATE devices SET ownerId = ?  WHERE deviceId = ? AND ownerId=?", newownerId, deviceid, oldownId)
	if err != nil {
		log.Println("所有者更新失败:", err)
		return err
	}
	err = updatefileowner(deviceid, oldownId, newownerId, db)
	if err != nil {
		log.Println("文件所有者更新失败:", err)
		return err
	}
	err = updateshareowner(deviceid, oldownId, newownerId, db)
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
	_, err = db.Exec("UPDATE sharing SET sharedOwerId = ?  WHERE deviceId = ? AND sharedOwerId=?", newownerid, deviceId, oldownerid)
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
//	@param devicesname 所属设备名
//	@param fileName 文件名
//	@param filePath 文件的储存路径
//	@param db 数据库连接对象
//	@return 没错误就返回nil
func AddFile(devicename string, fileName string, filePath string, db *sqlx.DB) error {
	uniqueID := uuid.New().String()
	ownerId, err := Querydevice(devicename, db, 1)
	if err != nil {
		log.Println("不存在该设备：", err)
		return err
	}
	devicesid, err := Querydevice(devicename, db, 0)
	query := `INSERT INTO files(deviceId,filePath,fileId,ownerId,fileName) VALUES (?,?,?,?,?)`
	_, err = db.Exec(query, devicesid, filePath, uniqueID, ownerId, fileName)
	if err != nil {
		log.Println("储存文件路径失败：", err)
		return err
	}
	fmt.Println("储存文件路径成功")
	return nil
}

// Function: QueryFile
//
//	Description: 查询
//
//	@param filename: 查询文件名
//	@param db: 链接数据库对象
//	@param params:分别返回0:文件id，1:设备id，2:所有者id和3:存储路径,请按顺序输入
//	@return string切片存放顺序与数字对应若正常则为nil
func QueryFile(filename string, db *sqlx.DB, params ...int) ([]string, error) {
	query := `SELECT fileId,ownerId,deviceId,filepath FROM files where fileName=?`

	type File struct {
		fileid   string `db:"fileId"`
		ownerid  string `db:"ownerId"`
		deviceid string `db:"deviceId""`
		filepath string `db:"filepath"`
	}
	var file File
	err := db.Get(&file, query, filename)
	if err != nil {
		println("查询出错或文件不存在")
		return nil, err
	}

	// 根据 params 参数返回相应字段
	var result []string
	for _, param := range params {
		switch param {
		case 0:
			result = append(result, file.fileid) // 返回 fileId
		case 1:
			result = append(result, file.ownerid) // 返回 ownerId
		case 2:
			result = append(result, file.deviceid) // 返回 deviceId
		case 3:
			result = append(result, file.filepath) // 返回 filepath
		default:
			return nil, fmt.Errorf("无效的参数: %d", param)
		}
	}

	return result, nil
}

// Function: DeleteFile
//
//	Description: 批量删除文件(以及共享表的删除)
//
//	@param filename: 查询文件名
//	@param db: 链接数据库对象
//	@param filenames:文件名
//	@return 若正常则为nil
func DeleteFile(db *sqlx.DB, filenames ...string) error {
	// 如果没有文件名传入，直接返回
	if len(filenames) == 0 {
		return nil
	}

	// 创建一个 IN 查询条件
	query, args, err := sqlx.In(`DELETE FROM files WHERE fileName IN (?)`, filenames)
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
	//删除共享表待补充
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
	password, err := queryPerson(ownername, db, 1)
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(inputpassword))
	if err != nil {
		log.Println("密码验证失败:", err)
		return err
	}
	return nil
}
func CheckDeviceName(ownerId string, deviceName string, db *sqlx.DB) {
	query := `SELECT deviceId FROM devices where deviceName=? AND ownerId=? `
	var deviceid string
	err := db.Get(&deviceid, query, deviceName, ownerId)
	if err != nil {
		fmt.Println("该设备名已存在，请修改名字")
	}
}
func CheckFileName(ownerId string, fileName string, db *sqlx.DB) {
	query := `SELECT fileId FROM files where fileName=? AND ownerId=? `
	var fileid string
	err := db.Get(&fileid, query, fileName, ownerId)
	if err != nil {
		fmt.Println("该文件名已存在，请修改名字")
	}
}
