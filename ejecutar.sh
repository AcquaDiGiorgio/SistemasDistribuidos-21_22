$(/usr/local/go/bin/go run server.go &)
echo $(/usr/local/go/bin/go run client.go $1 $2)