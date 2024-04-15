package application_controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_chat_model "github.com/thuanpham98/go-websocker-server/domain/chat/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
	"github.com/thuanpham98/go-websocker-server/infrastructure/protobuf/message_protobuf"
	"google.golang.org/protobuf/proto"
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

	go readPump(websocketConnection)

	go writePump(websocketConnection, userCh)
	
    for msg := range msgs {
		userCh <- msg.Body
	}
}

func SendMessageToGroup(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 400,
				Message: "Can not read body Data",
				Data: nil,
			},
		})
		return
	}

	var message message_protobuf.MessageRequest
	if err := proto.Unmarshal(body, &message); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 400,
				Message: "Can not read message",
				Data: nil,
			},
		})
		return
	}
	

	receiverId:= message.Receiver
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",receiverId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not find user",
				Data: nil,
			},
		})
		return
	}

	// save message into data base
	senderId,ok:=c.Get("user")
	messageType,checkMessageType:= domain_chat_model.ParseString(message.Type.String())
	if(!ok||!checkMessageType){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not find user",
				Data: nil,
			},
		})
		return
	}

	message_entity:=domain_chat_model.MessageEntity{
		ID: uuid.New().String(),
		Sender: senderId.(string),
		Receiver: receiverId,
		CreatedAt: time.Now().Format(time.RFC3339),
		Group: domain_chat_model.GroupEntity{
			Id: message.Group.Id,
			Name: message.Group.Name,
		},
		Content: message.Content,
		Type: messageType,
	}
	result:=infrastructure.DB.Create(&message_entity)

	if(result.Error!=nil){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Failed to save message",
				Data: nil,
			},
		})
		return
	}

	// ping message to cloud
	ch, err := infrastructure.MessageQueueConntection.Channel()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "False to create channel",
				Data: nil,
			},
		})
		return
	}
	defer ch.Close()
	ctxSender, cancelCtxSender := context.WithTimeout(context.Background(), 5000*time.Millisecond)

	defer cancelCtxSender()


	messageToCloud,errMessageToCloud:=proto.Marshal(&message_protobuf.MessageReponse{
		Id: message_entity.ID,
		Sender: message_entity.Sender,
		Receiver: message_entity.Receiver,
		Group: message.Group,
		Type: message.Type,
		CreateAt: message_entity.CreatedAt,
		Content: message_entity.Content,
	})

	if errMessageToCloud != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "False to noti message",
				Data: nil,
			},
		})
		return
	}

	errPushSender := ch.PublishWithContext(
		ctxSender,
		os.Getenv("EXCHANGE_NAME_CHAT_NEWS"), //exchange name
		"", // router key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/octet-stream",
			Body: messageToCloud,
		},
	)

	if  errPushSender!=nil {
		fmt.Println(errPushSender)
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "False to send message to user",
				Data: nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: true,
		Code: 0,
		Message: "success",
	}})
}
