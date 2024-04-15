package application_controllers

import (
	"context"
	"fmt"
	"io"
	"log"
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

// send message
func SendMessageToFriend(c *gin.Context) {
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
		Group: domain_chat_model.GroupEntity{},
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
	ctxReceiver, cancelCtxReceiver := context.WithTimeout(context.Background(), 5000*time.Millisecond)

	defer cancelCtxSender()
	defer cancelCtxReceiver()


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

	errPushReceiver := ch.PublishWithContext(
		ctxReceiver,
		os.Getenv("EXCHANGE_NAME_CHAT_POINT_TO_POINT"), //exchange name
		message.Receiver, // router key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/octet-stream",
			Body: messageToCloud,
		},
	)

	errPushSender := ch.PublishWithContext(
		ctxSender,
		os.Getenv("EXCHANGE_NAME_CHAT_POINT_TO_POINT"), //exchange name
		senderId.(string), // router key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/octet-stream",
			Body: messageToCloud,
		},
	)
	if errPushReceiver != nil || errPushSender!=nil {
		fmt.Println(errPushReceiver)
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

// listen message
func ListenMessageForUser(c *gin.Context){
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
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	
	// tạo channel để lắng nghe
	// infrastructure.MutexMessageChannels.Lock()
	var ch *amqp091.Channel
	ch, ok := infrastructure.MessageChannels[userid]
	if !ok {
		newCh, err := infrastructure.MessageQueueConntection.Channel()
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest,gin.H{
				"error":"Failed to create chanle",
			})
			return
		}
		ch=newCh
		infrastructure.MessageChannels[userid]=ch
	}
	// infrastructure.MutexMessageChannels.Unlock()
	
    err = ch.ExchangeDeclare(
            os.Getenv("EXCHANGE_NAME_CHAT_POINT_TO_POINT"), // Tên của exchange
            "direct",             // Loại của exchange
            true,                 // Durable
            false,                // Auto-deleted
            false,                // Internal
            false,                // No-wait
            nil,                  // Arguments
        )

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create chanle",
		})
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
		fmt.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create chanle",
		})
		return
	}

     // Ràng buộc hàng đợi với exchange
	err = ch.QueueBind(
		q.Name,       // Tên của hàng đợi
		userid,           // Routing key
		os.Getenv("EXCHANGE_NAME_CHAT_POINT_TO_POINT"), // Tên của exchange
		false,        // No-wait
		nil,          // Arguments
	)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create chanle",
		})
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
		fmt.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create chanle",
		})
		return
	}

	userCh := make(chan []byte)


	defer ch.Close()
    defer delete(infrastructure.MessageChannels,userid)
	defer close(userCh)
	defer websocketConnection.Close()

	go readPump(websocketConnection,func() {
		
		// close(userCh)
		websocketConnection.Close()
		ch.Close()
		delete(infrastructure.MessageChannels,userid)

		// websocketConnection.Close()
		// ch.Close()
		// delete(infrastructure.MessageChannels,userid)
	})

	go writePump(websocketConnection, userCh)

	
	
    for msg := range msgs {
		userCh <- msg.Body
		// select {
		// 	case userCh <- msg.Body:
		// 	default:
		// 		// Nếu không thể gửi tin nhắn, đóng kết nối WebSocket
		// 		websocketConnection.Close()
		// 		ch.Close()
		// 		close(userCh)
		// 		delete(infrastructure.MessageChannels,userid)
		// 	}
	}
}

func writePump(conn *websocket.Conn, userCh <-chan []byte) {
	for msg := range userCh {
		err := conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			log.Printf("Failed to write message to WebSocket: %v", err)
			break
		}
	}
}

func readPump(conn *websocket.Conn, closeCallback func()) {
	for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                log.Printf("WebSocket connection closed normally: %v", err)
            } else {
                log.Printf("Failed to read message from WebSocket: %v", err)
            }
			closeCallback()
            break // Thoát khỏi vòng lặp khi kết nối bị đóng
        }
    }
}

// get message with page
func GetMessagePageAble(c *gin.Context){
	userId,ok:=c.Get("user")
	if(!ok||userId==""){
		c.JSON(http.StatusBadRequest,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 105,
				Message: "user not found",
				Data: nil,
			},
		})
		return
	}

	var body struct{
		Page int `json:"page"`
		PageSize int `json:"page_size"`
	}
	fmt.Println(c.Request.Body)


	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not read body",
			Data: nil,
		}})
		return
	}

	var offset int = (body.Page) * body.PageSize

	var messages []domain_chat_model.MessageEntity
    if err:= infrastructure.DB.Raw("SELECT * FROM public.message_entities ORDER BY created_at DESC LIMIT ? OFFSET ?", body.PageSize, offset).Scan(&messages).Error; err != nil {
        c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not get data",
			Data: nil,
		}})
		return
    }

	var dtos []domain_chat_model.MessageDTO

    for _, mess := range messages {
        dto := domain_chat_model.MessageDTO{
			Id: mess.ID,
			Sender: mess.Sender,
			Receiver: mess.Receiver,
			Group: domain_chat_model.GroupDTO{
				Id: mess.Group.Id,
				Name: mess.Group.Name,
			},
			CreateAt: mess.CreatedAt,
			Content: mess.Content,
			Type: mess.Type,
        }
        dtos = append(dtos, dto)
    }
	
	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: dtos,
		Code: 0,
		Message: "success",
	}})
}