package application_controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context){
	var body struct{
		Name string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
		Phone string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read body",
		})
		return
	}

	hashed,err:=bcrypt.GenerateFromPassword([]byte(body.Password),bcrypt.DefaultCost)

	if(err!=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read password",
		})
		return
	}
	user:=domain_auth_model.UserEntity{
		Id: uuid.New().String(),
		Name: body.Name,
		Email: body.Email,
		Phone: body.Phone,
		Password: string(hashed),
	}
	result:=infrastructure.DB.Create(&user)

	if(result.Error!=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create account",
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{"susscess": true})
}

func Login(c *gin.Context){
	var body struct{
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not read body",
			Data: nil,
		}})
		return
	}

	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"phone = ?",body.Username)
	
	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{"data": domain_common_model.CommonReponse{
			Code: 105,
			Message: "user not found",
			Data: nil,
		}})
		return
	}

	err:=bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(body.Password))

	if(err!=nil){
		c.JSON(http.StatusNotFound,gin.H{"data": domain_common_model.CommonReponse{
			Code: 105,
			Message: "user not found",
			Data: nil,
		}})
		return
	}

	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"sub":user.Email,
		"exp":time.Now().Add(time.Hour*24).Unix(),
	})

	tokenString,err:= token.SignedString([]byte(os.Getenv("JWT_HMAC_SECRET_KEY")))

	if(err!=nil){
		c.JSON(http.StatusNotFound,gin.H{"data": domain_common_model.CommonReponse{
			Code: 105,
			Message: "login error",
			Data: nil,
		}})
		return
	}

	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("Authorization",tokenString,3600*24,"","",false,true)

	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Code: 0,
		Message: "success",
		Data: tokenString,
	}})
}