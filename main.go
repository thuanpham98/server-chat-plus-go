package main

import (
	"os"

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
	infrastructure.ConnectToDatabase()
    infrastructure.SyncDatabase()
    infrastructure.ConnectMessageQueue()
    infrastructure.MessageChannels=make(map[string]*amqp091.Channel)
    infrastructure.ConnectStorage()

    defer infrastructure.MessageQueueConntection.Close()

	//we need pass hub to out route with roomid
    gin.SetMode(os.Getenv("GIN_MODE"))
	app := gin.Default()
    corCf:=cors.DefaultConfig()
    corCf.AllowHeaders=[]string{"Authorization", "Content-Type","Cookie"}
    corCf.AllowOrigins = []string{"http://localhost:8080","http://chat-app.com.vn"}
    corCf.AllowCredentials = true
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
    user.GET("/info",application_middlewares.ValidateTokenCoockie,application_controllers.GetUserInfo)
    user.GET("/friends",application_middlewares.ValidateTokenCoockie,application_controllers.GetListFriends)

    // message router
    message:= routerVersion.Group("/message")
    message.POST("/send",application_middlewares.ValidateTokenCoockie,application_controllers.SendMessageToFriend)
    message.POST("/send-to-group",application_middlewares.ValidateTokenCoockie,application_controllers.SendMessageToGroup)
    message.POST("/list",application_middlewares.ValidateTokenCoockie,application_controllers.GetMessagePageAble)
    message.POST("/delete",application_middlewares.ValidateTokenCoockie,application_controllers.DeleteMessage)


    // group router
    group:= routerVersion.Group("/group")
    group.POST("/create",application_middlewares.ValidateTokenCoockie,application_controllers.CreateGroup)


    // storage router
    storage:=routerVersion.Group("/storage")
    storage.POST("/upload",application_middlewares.ValidateTokenCoockie,application_controllers.UploadFile)
    storage.POST("/download",application_middlewares.AuthRequire,application_controllers.DownloadFile)
    storage.GET("/public/image/:image",application_middlewares.ValidateTokenCoockie,application_controllers.PreviewImage)


    // websocket router
	app.GET("user/ws/message",application_middlewares.ValidateTokenCoockie,func(c *gin.Context) {
        application_controllers.ListenMessageForUser(c)
	})

    app.GET("/group/ws/:groupid",application_middlewares.ValidateTokenCoockie,application_controllers.ListenMessageFromGroup)

	app.Run()
    // <-forever
}