package initializers

import (
	"os"

	"github.com/rabbitmq/amqp091-go"
	application_errors "github.com/thuanpham98/go-websocker-server/application/errors"
)

var MessageQueueConntection *amqp091.Connection

func ConnectMessageQueue (){
	conn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	MessageQueueConntection=conn
    application_errors.FailOnError(err, "Failed to connect to RabbitMQ")
}