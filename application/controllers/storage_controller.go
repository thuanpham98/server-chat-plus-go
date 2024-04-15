package application_controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)

func UploadFile(c *gin.Context){
	userId,ok:=c.Get("user")
	if(!ok){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 403,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",userId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 403,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}

	file, errReadBody := c.FormFile("file")

	if errReadBody != nil {
		fmt.Println(c.Request.Form)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not read body",
				Data: nil,
			},
		})
		return
	}

	src, errReadFile := file.Open()
	if errReadFile != nil {
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not read file",
				Data: nil,
			},
		})
		return
	}
	defer src.Close()

	// Tải tệp lên MinIO
	fmt.Println(file.Filename)
	fmt.Println(file.Size)

	info, errUploadFile := infrastructure.MinIOClient.PutObject(context.Background(), "file", file.Filename, src, file.Size, minio.PutObjectOptions{})

	if errUploadFile != nil {
		log.Fatalln(errUploadFile)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not upload file",
				Data: nil,
			},
		})
		return
	}

	// presignedURL, errPresigned :=infrastructure.MinIOClient.PresignedGetObject(context.Background(),"file",info.Key,time.Hour*24,nil)

	// if errPresigned != nil {
	// 	log.Fatalln(errUploadFile)
	// 	c.JSON(http.StatusNotFound,gin.H{
	// 		"error": domain_common_model.CommonReponse{
	// 			Code: 404,
	// 			Message: "Can not upload file",
	// 			Data: nil,
	// 		},
	// 	})
	// 	return
	// }
	// fmt.Println(presignedURL)
	
	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: info.Key,
		Code: 0,
		Message: "success",
	}})
}

func DownloadFile(c *gin.Context){
	userId,ok:=c.Get("user")
	if(!ok){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",userId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}

	var body struct{
		FileName string `json:"file_name" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
			c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Not read file name",
				Data: nil,
			},
		})
		return
	}

	file,errGetFile := infrastructure.MinIOClient.GetObject(context.Background(), "file", body.FileName, minio.GetObjectOptions{})
	if errGetFile != nil {
		log.Fatalln(errGetFile)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not read body",
				Data: nil,
			},
		})
		return
	}

		// Thiết lập tiêu đề của phản hồi HTTP
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+body.FileName)
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		// Sao chép nội dung của đối tượng MinIO vào phản hồi HTTP
		if _, err := io.Copy(c.Writer, file); err != nil {
			log.Fatalln(err)
		}
}

func PreviewFile(c *gin.Context){
	userId,ok:=c.Get("user")
		fmt.Println(ok)

	if(!ok){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",userId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Not found",
				Data: nil,
			},
		})
		return
	}

	fileParam := c.Param("file")

	file,errGetFile := infrastructure.MinIOClient.GetObject(context.Background(), "file", fileParam, minio.GetObjectOptions{})


	if errGetFile != nil {
		log.Fatalln(errGetFile)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not read body",
				Data: nil,
			},
		})
		return
	}

		// Thiết lập tiêu đề của phản hồi HTTP
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+ fileParam)
		c.Writer.Header().Set("Content-Type", "image/jpeg")

		// Sao chép nội dung của đối tượng MinIO vào phản hồi HTTP
		if _, err := io.Copy(c.Writer, file); err != nil {
			log.Fatalln(err)
		}
}