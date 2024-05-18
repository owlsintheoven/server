package main

import (
	"log"
	"os"
	"owlsintheoven/learning-go/tcp/server/listeners"
)

const (
	SERVER_TYPE = "tcp"
	SERVER_HOST = "localhost"
	SERVER_PORT = "9000"
)

func main() {
	serverType := os.Args[1]
	log.Println("Server Running...")
	switch serverType {
	case "1":
		listeners.ProcessAConnection(SERVER_TYPE, SERVER_HOST, SERVER_PORT)
	case "2":
		listeners.ProcessConnectionsWithThreads(SERVER_TYPE, SERVER_HOST, SERVER_PORT)
	case "4":
		listeners.ProcessAndSendFixedMessages(SERVER_TYPE, SERVER_HOST, SERVER_PORT)
	case "5":
		listeners.ProcessConnectionsAndBroadcast(SERVER_TYPE, SERVER_HOST, SERVER_PORT)
	case "6":
		listeners.ProcessChatGroup(SERVER_TYPE, SERVER_HOST, SERVER_PORT)
	case "7":
		listeners.SOCKS4(SERVER_TYPE, SERVER_HOST, "1080")
	default:
		log.Fatalln("Undefined")
	}

	log.Println("Server stopped")
}
