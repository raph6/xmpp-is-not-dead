package main

import (
	"log"

	"github.com/raph6/xmpp-is-not-dead/server"
)

func main() {
	address := "localhost:5222"

	srv := server.NewServer(address)
	log.Printf("Starting XMPP server on %s", address)
	srv.Start()
}
