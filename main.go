package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	application_controllers "github.com/thuanpham98/go-websocker-server/application/controllers"
	application_middlewares "github.com/thuanpham98/go-websocker-server/application/middlewares"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)

func main() {
	// initialize
	infrastructure.LoadEnvVariable()
	infrastructure.ConnectToPostgresql()
    infrastructure.SyncDatabase()
    infrastructure.ConnectMessageQueue()
    infrastructure.MessageChannels=make(map[string]*amqp091.Channel)

    defer infrastructure.MessageQueueConntection.Close()


	// config for websocket 
	// hub := chat.NewHub()
	// go hub.Run()

	//we need pass hub to out route with roomid
	app := gin.Default()
    corCf:=cors.DefaultConfig()
    corCf.AllowHeaders=[]string{"Authorization", "Content-Type"}
    corCf.AllowOrigins = []string{"*"} 
    // cors
    app.Use(cors.New(corCf))
    // version router
    routerVersion := app.Group("/api/v1")
    // auth router
    auth :=routerVersion.Group("/auth")
    auth.POST("/signup",application_controllers.SignUp)
    auth.POST("/login",application_controllers.Login)
    // user router
    user:= routerVersion.Group("/user")
    user.GET("/info",application_middlewares.AuthRequire,application_controllers.GetUserInfo)
    user.GET("/friends",application_middlewares.AuthRequire,application_controllers.GetListFriends)

    // message router
    message:= routerVersion.Group("/message")
    message.POST("/send",application_middlewares.AuthRequire,application_controllers.SendMessageToFriend)

    // websocket router
	app.GET("user/ws/message",application_middlewares.ValidateTokenCoockie,func(c *gin.Context) {
		// roomId := c.Param("roomId")
		// chat.ServeWS(c, roomId, hub)
        application_controllers.ListenMessageForUser(c)
        // 
	})

	app.Run()
    // <-forever
}