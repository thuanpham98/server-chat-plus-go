```
PORT=6969
DBCONFIG="host=localhost user=postgres password=1 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Bangkok"
RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
GIN_MODE=debug
# release
# debug
PASSWORD_HASH_SECRET_KEY=ahihi
JWT_HMAC_SECRET_KEY=ahuhu
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY_ID=8vjT1cKZ6vFZFdBd2tBn
MINIO_SECRET_ACCESS_KEY=22AbXf8Hh4hjX0xGSb8Kih5cPNQgnqtbGTprYFgt

EXCHANGE_NAME_CHAT_POINT_TO_POINT="amq.direct"
EXCHANGE_NAME_CHAT_NEWS="amq.fanout"
```

```
protoc --proto_path=infrastructure/protobuf --go_out=infrastructure infrastructure/protobuf/message.proto
```

build

```
go build -o bin/chat-app && cp .env bin
```
