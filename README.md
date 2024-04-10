```
PORT=6969
DBCONFIG="host=localhost user=postgres password=1 dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Bangkok"
RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
GIN_MODE=debug
# release
PASSWORD_HASH_SECRET_KEY=ahihi
JWT_HMAC_SECRET_KEY=ahuhu
```

```
protoc --proto_path=infrastructure/protobuf --go_out=infrastructure infrastructure/protobuf/message.proto
```
