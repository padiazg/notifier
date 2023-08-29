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