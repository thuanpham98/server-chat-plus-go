package application_controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)

func GetUserInfo(c *gin.Context){
	userId,ok:=c.Get("user")
	if(!ok){
		c.JSON(http.StatusNotFound,gin.H{"error": "user not found"})
		return
	}
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",userId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{"error": "user not found"})
		return
	}
	
	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: domain_auth_model.UserDTO{
			Name: user.Name,
			Email: user.Email,
			Id: user.Id,
			Phone: user.Phone,
		},
		Code: 0,
		Message: "success",
	}})
}

func GetListFriends(c *gin.Context){
	userId,ok:=c.Get("user")
	if(!ok){
		c.JSON(http.StatusNotFound,gin.H{"error": "user not found"})
		return
	}
	var users []domain_auth_model.UserEntity
	result := infrastructure.DB.Where("Id <> ?", userId).Find(&users)

	if(result.Error!=nil){
		c.JSON(http.StatusNotFound,gin.H{"error": "Query list friend error"})
		return
	}

	var dtos []domain_auth_model.UserDTO

    for _, user := range users {
        dto := domain_auth_model.UserDTO{
			Id: user.Id,
			Name: user.Name,
			Email: user.Email,
			Phone: user.Phone,
        }
        dtos = append(dtos, dto)
    }
	
	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: dtos,
		Code: 0,
		Message: "success",
	}})
}