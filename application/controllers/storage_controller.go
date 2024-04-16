package application_controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

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
		fmt.Printf("Error: %v",errReadBody)
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

		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+body.FileName)
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		if _, err := io.Copy(c.Writer, file); err != nil {
			log.Fatalln(err)
		}
}

func PreviewImage(c *gin.Context){
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

	fileParam := c.Param("image")

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

	defer file.Close()

	parts := strings.Split(fileParam, ".")
	extension := parts[len(parts)-1]

	mimeType := mime.TypeByExtension("." + extension)

	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+ fileParam)
	c.Writer.Header().Set("Content-Type", mimeType)

	if _, err := io.Copy(c.Writer, file); err != nil {
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not fine file",
				Data: nil,
			},
		})
	}
}