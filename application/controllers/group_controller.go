package application_controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	domain_chat_model "github.com/thuanpham98/go-websocker-server/domain/chat/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)


func CreateGroup(c *gin.Context){
	var body struct{
		Name string `json:"name" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read body",
		})
		return
	}


	group:=domain_chat_model.GroupEntity{
		Id: uuid.New().String(),
		Name: body.Name,
	}

	result:=infrastructure.DB.Create(&group)

	if(result.Error!=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 400,
				Message: "Can not create group",
				Data: nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Code: 0,
		Message: "Success",
		Data: true,
	},})
}

func ListenMessageFromGroup(c *gin.Context){
	groupid:=c.Param("groupid")

	retContext,okCheckUserId := c.Get("user")
	if(!okCheckUserId){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"user not found",
		})
		return
	}

	userid:=retContext.(string)

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024*512,
		CheckOrigin:func(r *http.Request) bool { return true } ,
	}
	websocketConnection, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Failed to upgrade connection to WebSocket:", err)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}
	
	// tạo channel để lắng nghe
	// infrastructure.MutexMessageChannels.Lock()
	var ch *amqp091.Channel
	ch, ok := infrastructure.MessageChannels[groupid]
	if !ok {
		newCh, err := infrastructure.MessageQueueConntection.Channel()
		if err != nil {
			fmt.Println("Failed to upgrade connection to WebSocket:", err)
			c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
				Code: 400,
				Message: "Can not connect websocket",
				Data: nil,
			}})
			return
		}
		ch=newCh
		infrastructure.MessageChannels[groupid]=ch
	}
	// infrastructure.MutexMessageChannels.Unlock()
	
    err = ch.ExchangeDeclare(
            os.Getenv("EXCHANGE_NAME_CHAT_NEWS"), // Tên của exchange
            "fanout",             // Loại của exchange
            true,                 // Durable
            false,                // Auto-deleted
            false,                // Internal
            false,                // No-wait
            nil,                  // Arguments
        )

	if err != nil {
		fmt.Println("Failed to upgrade connection to WebSocket:", err)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}

    q, err := ch.QueueDeclare(
        userid, // Tên hàng đợi
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )

    if err != nil {
		fmt.Println("Failed to upgrade connection to WebSocket:", err)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}

     // Ràng buộc hàng đợi với exchange
	err = ch.QueueBind(
		q.Name,       // Tên của hàng đợi
		"",           // Routing key
		os.Getenv("EXCHANGE_NAME_CHAT_NEWS"), // Tên của exchange
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		fmt.Println("Failed to upgrade connection to WebSocket:", err)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )

    if err != nil {
		fmt.Println("Failed to upgrade connection to WebSocket:", err)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}

	userCh := make(chan []byte)

	defer ch.Close()
    defer delete(infrastructure.MessageChannels,groupid)
	defer close(userCh)
	defer websocketConnection.Close()

	go readPump(websocketConnection,func() {
		close(userCh)
		websocketConnection.Close()
		ch.Close()
		delete(infrastructure.MessageChannels,groupid)
	})

	go writePump(websocketConnection, userCh)
	
    for msg := range msgs {
		userCh <- msg.Body
	}
}
