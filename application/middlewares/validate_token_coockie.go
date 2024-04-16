package application_middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)

func ValidateTokenCoockie(c *gin.Context){
	tokenString,err:=c.Cookie("Authorization");

	if err!=nil {
		fmt.Printf("No more coockie %v",err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	retToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_HMAC_SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := retToken.Claims.(jwt.MapClaims); ok {
		if(float64(time.Now().Unix()) > claims["exp"].(float64)){
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userEmail:= claims["sub"]
		var user domain_auth_model.UserEntity
		infrastructure.DB.First(&user,"email = ?",userEmail)
		if(user.Id==""){
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("user",user.Id)
		c.Next()

	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}