package application_controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	"github.com/thuanpham98/go-websocker-server/initializers"
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

	hashed,err:=bcrypt.GenerateFromPassword([]byte(body.Password),8)

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
	result:=initializers.DB.Create(&user)

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
		Phone string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read body",
		})
		return
	}

	var user domain_auth_model.UserEntity
	initializers.DB.First(&user,"phone = ?",body.Phone)

	
	if(user.Id==""){

		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Invalide user name or password",
		})
		return
	}

	err:=bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(body.Password))

	if(err!=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Invalide user name or password",
		})
		return
	}

	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"sub":user.Email,
		"exp":time.Now().Add(time.Hour*24).Unix(),
	})

	tokenString,err:= token.SignedString([]byte(os.Getenv("PASSWORD_SECRET")))

	if(err!=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Invalide user name or password",
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{"token": tokenString})
}