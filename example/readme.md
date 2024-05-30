# Notifier example

## AMQP receiver
### Run RabbitMQ
```bash
docker compose --env-file=.env up -d
```

## Webhook receiver
### Create SSL cert
```bash
$ make cert
$ openssl req -x509 \
-newkey rsa:4096 \
-days 365 \
-nodes \
-subj "/CN=localhosr" \
-addext "subjectAltName=DNS:localhost" \
-keyout cert/localhost.key \
-out cert/localhost.crt
```

## Run the demo
```bash
# run MQ receiver
go run amqp-receiver/amqp-receiver.go

# run webhook receiver
go run webhook-receiver/webhook-receiver.go

# run emmiter
go run emitter.go
```