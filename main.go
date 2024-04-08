package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	application_controllers "github.com/thuanpham98/go-websocker-server/application/controllers"
	application_errors "github.com/thuanpham98/go-websocker-server/application/errors"
	application_middlewares "github.com/thuanpham98/go-websocker-server/application/middlewares"
	"github.com/thuanpham98/go-websocker-server/chat"
	"github.com/thuanpham98/go-websocker-server/initializers"
)

func main() {
	// initialize
	initializers.LoadEnvVariable()
	initializers.ConnectToPostgresql()
    initializers.SyncDatabase()
    initializers.ConnectMessageQueue()

    defer initializers.MessageQueueConntection.Close()


    ch, err := initializers.MessageQueueConntection.Channel()
     application_errors.FailOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "hello", // Tên hàng đợi
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
     application_errors.FailOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
     application_errors.FailOnError(err, "Failed to register a consumer")

    // forever := make(chan bool)

    // Set Context
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // test send message
    body := "Hello World!"
    err = ch.PublishWithContext( // errors here
       ctx,    // context
       "",     // exchange
       q.Name, // routing key
       false,  // mandatory
       false,  // immediate
       amqp.Publishing{
          ContentType: "text/plain",
          Body:        []byte(body),
       })
	    application_errors.FailOnError(err,"error send body")

    go func() {
        for d := range msgs {
            log.Printf("Received a message: %s", d.Body)
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	// config for websocket 
	hub := chat.NewHub()
	go hub.Run()

	//we need pass hub to out route with roomid
	app := gin.Default()
    routerVersion := app.Group("/api/v1")

    // auth
    auth :=routerVersion.Group("/auth")
    auth.POST("/signup",application_controllers.SignUp)
    auth.POST("/login",application_controllers.Login)


    // user
    user:= routerVersion.Group("/user")
    user.GET("/:id",application_middlewares.AuthRequire,application_controllers.GetUserInfo)

    // websocket
	app.GET("/ws/:roomId",func(c *gin.Context) {
        fmt.Println("connect")
		roomId := c.Param("roomId")
		chat.ServeWS(c, roomId, hub)
	})

	app.Run()
    // <-forever


}