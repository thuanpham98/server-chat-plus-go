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
	// kiểm tra user và group có tồn tại hay không
	groupId:=c.Param("groupid")
	var group domain_chat_model.GroupEntity
	infrastructure.DB.First(&group,"id = ?",groupId)
	if(group.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "group not found",
				Data: nil,
			},
		})
		return
	}
	retContext,okCheckUserId := c.Get("user")
	if(!okCheckUserId){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 105,
				Message: "group not found",
				Data: nil,
			},
		})
		return
	}
	userid:=retContext.(string)

	// upgrate lên websocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024*1024,
		CheckOrigin:func(r *http.Request) bool { return true } ,
	}
	wsConn, errUpgradeWs := upgrader.Upgrade(c.Writer, c.Request, nil)
	if errUpgradeWs != nil {
		fmt.Printf("Failed to upgrade connection to WebSocket: %v", errUpgradeWs)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Không thể khởi tạo websocket",
			Data: nil,
		}})
		return
	}
	defer wsConn.Close()
	
	// tạo channel để lắng nghe
	rabbChannel, errCreateChannel := infrastructure.MessageQueueConntection.Channel()
	if errCreateChannel != nil {
		fmt.Printf("Failed to upgrade connection to WebSocket: %v", errCreateChannel)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Không thể khởi tạo websocket",
			Data: nil,
		}})
		return
	}
	defer rabbChannel.Close()

	// khai báo queue
    rabbQueue, errCreateQueue := rabbChannel.QueueDeclare(
        userid+"-"+groupId, // Tên hàng đợi
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    if errCreateQueue != nil {
		fmt.Printf("Failed to upgrade connection to WebSocket: %v", errCreateQueue)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}
	defer rabbChannel.QueueDelete(rabbQueue.Name,false,false,true)

	// Ràng buộc hàng đợi với exchange
	errBindingQueue := rabbChannel.QueueBind(
		rabbQueue.Name,       // Tên của hàng đợi
		"",           // Routing key
		os.Getenv("EXCHANGE_NAME_CHAT_NEWS"), // Tên của exchange
		false,        // No-wait
		nil,          // Arguments
	)
	if errBindingQueue != nil {
		fmt.Printf("Failed to upgrade connection to WebSocket: %v", errBindingQueue)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}
	defer rabbChannel.QueueUnbind(rabbQueue.Name,"",os.Getenv("EXCHANGE_NAME_CHAT_NEWS"),amqp091.Table{})

	// khởi tạo consumer để lắng nghe message
    rabbConsumer, errConsume := rabbChannel.Consume(
        rabbQueue.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    if errConsume != nil {
		fmt.Printf("Failed to upgrade connection to WebSocket: %v", errConsume)
		c.JSON(http.StatusBadRequest,gin.H{"data": domain_common_model.CommonReponse{
			Code: 400,
			Message: "Can not connect websocket",
			Data: nil,
		}})
		return
	}

	// userCh := make(chan []byte)
	// defer close(userCh)

	closeCh := make(chan struct{})

	// Lắng nghe sự kiện từ amqp091.Delivery
	go func() {
		for delivery := range rabbConsumer {
			// Xử lý dữ liệu từ AMQP và gửi qua WebSocket
			errWriteMessage := wsConn.WriteMessage(websocket.BinaryMessage, delivery.Body)
			if errWriteMessage != nil {
				fmt.Printf("Failed to write WebSocket: %v", errWriteMessage)
			}
		}
	}()

	// Lắng nghe sự kiện từ WebSocket
	go func() {
		defer close(closeCh)
		for {
			_, _, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					fmt.Printf("WebSocket connection closed normally: %v", err)
				} else {
					fmt.Printf("Failed to read message from WebSocket: %v", err)
				}
				break // Thoát khỏi vòng lặp khi kết nối bị đóng
			}
		}
	}()

	// Đóng kết nối khi có sự kiện từ WebSocket hoặc AMQP
	x, ok := <-closeCh
	fmt.Printf("WebSocket connection closed %v %v",x,ok)
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
	

	groupId:= message.Group.Id
	var group domain_chat_model.GroupEntity
	infrastructure.DB.First(&group,"id = ?",groupId)

	if(group.Id==""){
		c.JSON(http.StatusNotFound,gin.H{
			"error": domain_common_model.CommonReponse{
				Code: 404,
				Message: "Can not find group",
				Data: nil,
			},
		})
		return
	}

	// save message into data base
	senderId,ok:=c.Get("user")
	if(!ok){
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
		CreatedAt: time.Now().Format(time.RFC3339),
		Group: domain_chat_model.GroupEntity{
			Id: message.Group.Id,
			Name: message.Group.Name,
		},
		Content: message.Content,
		Type: domain_chat_model.MessageType(message.Type),
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
