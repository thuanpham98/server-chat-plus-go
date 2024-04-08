package application_controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	"github.com/thuanpham98/go-websocker-server/initializers"
)

func GetUserInfo(c *gin.Context){
	userId:= c.Param("id")
	var user domain_auth_model.UserEntity
	initializers.DB.First(&user,"id = ?",userId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{"error": "user not found"})
		return
	}
	
	c.JSON(http.StatusOK,gin.H{"data": domain_auth_model.UserDTO{
		Name: user.Name,
		Email: user.Email,
		Id: user.Id,
		Phone: user.Phone,
	}})
}