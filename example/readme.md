```go
$ go run webhook-client.go
$ go run emitter.go
```

To generate the TLS certificates, run the following command:
```shell
$ openssl req -x509 \                                  
-newkey rsa:4096 \     
-days 365 \
-nodes \                    
-subj "/CN=localhost" \  
-addext "subjectAltName=DNS:localhost" \
-keyout localhost.key \
-out localhost.crt
```

# MQ Queue

## RabbitMQ
### Install
```shell
docker run --rm -d \
--hostname my-rabbit \
--name rabbitmq \
--mount type=bind,source="$(pwd)"/rabbitmq/enabled-plugins,target=/etc/rabbitmq/enabled_plugins \
-p 8090:15672 \
-p 5672:5672 \
rabbitmq:3-alpine
```

### Get MQ nodes
```shell
curl -s \
-u guest:guest \
-H "content-type:application/json" \
http://localhost:8090/api/nodes | jq
[
  {
    "name": "rabbit@my-rabbit",
    "type": "disc",
    "running": true,
    "being_drained": false
  }
]
```

RabbitMQ Management API Reference: https://rawcdn.githack.com/rabbitmq/rabbitmq-server/v3.12.4/deps/rabbitmq_management/priv/www/api/index.html

### Create MQ queue
```shell
# get node name
curl -s \
-u guest:guest \
-H "content-type:application/json" \
http://localhost:8090/api/nodes | jq ".[0].name"

# "notifier" at the url is the queue name
# replace "node" value with the output of the above command
curl -s \
-u guest:guest \
-X PUT \
-H "content-type:application/json" \
http://localhost:8090/api/queues/%2F/notifier \
--data-binary '{"auto_delete":false,"durable":true,"arguments":{},"node":"rabbit@my-rabbit"}'
```

### Publish a test message
```shell
curl -s \
-u guest:guest \
-H "content-type:application/json" \
http://localhost:8090/api/exchanges/%2F/amq.default/publish \
--data-binary '{"vhost":"/","name":"amq.default","properties":{"delivery_mode":1,"headers":{}},"routing_key":"notifier","delivery_mode":"1","payload":"This is a test message","headers":{},"props":{},"payload_encoding":"string"}'
```
