package controllers

import (
	"MYSQL/database"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Fileid struct {
	Fileid   string `json:"fileid"`
	Name     string `json:"name"`
	FilePath string `json:"filePath" `
	DeviceId string `json:"deviceid" `
	OwnerId  string `json:"owner_id "`
}

var minioClient *minio.Client
var endpoint = "play.min.io"
var accessKeyID = "YOUR-ACCESSKEYID"
var secretAccessKey = "YOUR-SECRETACCESSKEY"
var useSSL = true
var bucketName = "my-bucketname"

// 需要名字，设备id，所属者id,原系统路径
func AddFile(c *gin.Context) {
	type addfile struct {
		Name     string `json:"name" binding:"required"`
		DeviceId string `json:"deviceid" `
		OwnerId  string `json:"owner_id " binding:"required"`
		FilePath string `json:"filePath" binding:"required"`
	}
	var Id addfile
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	Core, err := initCoreClient()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filepath, err := multipartUpload(Core, bucketName, Id.Name, Id.FilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.AddFile(Id.DeviceId, Id.Name, Id.OwnerId, filepath, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

// 需要文件id
func GetFilePath(c *gin.Context) {
	var Id Fileid
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	filepath, err := database.QueryFile(Id.DeviceId, db, 1)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"filePath": filepath})
}

// 需要提供设备Id和所有者id
func GetALLFile(c *gin.Context) {
	// 获取路径参数 :page
	pageParam := c.Param("page")
	// 将 page 参数转换为整数
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		// 如果转换失败，返回错误响应
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	var Id Fileid
	err = c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}

	personInfo, err := database.QueryAllFile(Id.DeviceId, Id.OwnerId, (page-1)*10, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"files": personInfo,
	})
}

// 文件id,新设备id，旧拥有者id，新拥有者id,其中旧拥有者Id为URL的一部分
func UpdateFileOwner(c *gin.Context) {
	oldownerid := c.Param("id")
	var Id Fileid
	err := c.ShouldBindJSON(&Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	err = database.ChangeFileOwn(Id.Fileid, Id.DeviceId, oldownerid, Id.OwnerId, db)
	if err != nil {
		log.Fatalln(err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "device updated", "fileId": Id.Fileid})
}

// 可以提供单个文件id来删除
func DeleteFile(c *gin.Context) {
	fileid := c.Param("id")
	db, err := database.ConnectMysql()
	if err != nil {
		log.Fatalln("连接失败")
	}
	name, err := database.QueryFile(fileid, db, 0)
	objectName := name[0]
	if err != nil {
		log.Fatalln("读取名字失败")
	}
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	err = minioClient.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	err = database.DeleteFile(db, fileid)
	if err != nil {
		log.Fatalln(err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "device deleted", "fileId": fileid})
}

func initCoreClient() (*minio.Core, error) {
	coreClient, err := minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL, // 如果使用 HTTPS，改为 true
	})
	if err != nil {
		return nil, err
	}
	return coreClient, nil
}

// 执行 Multipart Upload
func multipartUpload(core *minio.Core, bucketName, objectName, filePath string) (string, error) {
	// 打开需要上传的文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	// 检查存储桶是否存在
	exists, err := core.BucketExists(context.Background(), bucketName)
	if err != nil {
		return "", fmt.Errorf("could not check bucket: %v", err)
	}
	if !exists {
		return "", fmt.Errorf("bucket %s does not exist", bucketName)
	}

	// 开始 Multipart Upload
	uploadID, err := core.NewMultipartUpload(context.Background(), bucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("could not initiate multipart upload: %v", err)
	}

	// 分块上传
	var partNumber int
	var parts []minio.CompletePart
	buffer := make([]byte, 5*1024*1024) // 每次上传 5MB
	ctx := context.Background()
	for {
		// 读取文件块
		readBytes, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("could not read file: %v", err)
		}
		if readBytes == 0 {
			break
		}

		partNumber++
		partReader := io.NewSectionReader(file, 0, int64(readBytes))

		// 上传文件块
		etag, err := core.PutObjectPart(ctx, bucketName, objectName, uploadID, partNumber, partReader, int64(readBytes), minio.PutObjectPartOptions{})
		if err != nil {
			return "", fmt.Errorf("could not upload part %d: %v", partNumber, err)
		}

		// 记录已上传的块信息
		parts = append(parts, minio.CompletePart{
			PartNumber: partNumber,
			ETag:       etag.ETag,
		})
	}

	// 完成 Multipart Upload
	Info, err := core.CompleteMultipartUpload(
		context.Background(),
		bucketName,
		objectName,
		uploadID,
		parts,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("could not complete multipart upload: %v", err)
	}

	fmt.Println("Multipart upload completed successfully!")
	return Info.Location, nil
}
