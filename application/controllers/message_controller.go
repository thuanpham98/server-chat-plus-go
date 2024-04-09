package application_controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_chat_model "github.com/thuanpham98/go-websocker-server/domain/chat/model"
	domain_common_model "github.com/thuanpham98/go-websocker-server/domain/common/model"
	"github.com/thuanpham98/go-websocker-server/infrastructure"
)

func SendMessageToFriend(c *gin.Context) {
	receiverId:= c.Param("id")
	var user domain_auth_model.UserEntity
	infrastructure.DB.First(&user,"id = ?",receiverId)

	if(user.Id==""){
		c.JSON(http.StatusNotFound,gin.H{"error": "user not found"})
		return
	}

	var body struct{
		To string `json:"to" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if(c.Bind(&body) !=nil){
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read body",
		})
		return
	}

	ch, err := infrastructure.MessageQueueConntection.Channel()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to create chanle",
		})
		return
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	err = ch.PublishWithContext(
		ctx,
		body.To, //exchange name
		body.To, // router key
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body: []byte(body.Content),
		},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error":"Failed to read body",
		})
		return
	}

	senderId,ok:=c.Get("user")
	if(ok){
		message:=domain_chat_model.MessageEntity{
			ID: uuid.New().String(),
			Sender: senderId.(string),
			Receiver: receiverId,
			CreatedAt: time.Now(),
			Group: domain_chat_model.GroupEntity{},
			Content: body.Content,
		}
		result:=infrastructure.DB.Create(&message)

		if(result.Error!=nil){
			c.JSON(http.StatusBadRequest,gin.H{
				"error":"Failed to create message",
			})
			return
		}
	}

	c.JSON(http.StatusOK,gin.H{"data": domain_common_model.CommonReponse{
		Data: true,
		Code: 0,
		Message: "success",
	}})
	

}

func ListenMessageForUser(c *gin.Context){
	userid := c.Param("userid")
    log.Printf(" [*] %v Waiting for messages. To exit press CTRL+C",userid)


		// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:func(r *http.Request) bool { return true } ,
	}
	websocketConnection, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	
	// tạo channel để lắng nghe
	infrastructure.MutexMessageChannels.Lock()
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
	infrastructure.MutexMessageChannels.Unlock()
	
    err = ch.ExchangeDeclare(
            userid, // Tên của exchange
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
		userid, // Tên của exchange
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
		websocketConnection.Close()
		ch.Close()
		delete(infrastructure.MessageChannels,userid)
	})

	go writePump(websocketConnection, userCh)

	
	
    for msg := range msgs {
		// userCh <- msg.Body
		select {
			case userCh <- msg.Body:
			default:
				// Nếu không thể gửi tin nhắn, đóng kết nối WebSocket
				websocketConnection.Close()
				ch.Close()
				close(userCh)
				delete(infrastructure.MessageChannels,userid)
			}
	}
}

func writePump(conn *websocket.Conn, userCh <-chan []byte) {
	for msg := range userCh {
		err := conn.WriteMessage(websocket.TextMessage, msg)
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
