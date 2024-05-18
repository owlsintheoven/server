package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"owlsintheoven/learning-go/common"
	"sync"
	"syscall"
)

const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"
)

func main() {
	// establish connection
	port := os.Args[1]
	connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+port)
	if err != nil {
		log.Println("Error connecting: ", err.Error())
		return
	}

	// implement graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		common.WaitForShutdown(sigs)
		cancel()
	}()
	handle(ctx, connection)
	log.Println("Client stopped")
}

func handle(ctx context.Context, connection net.Conn) {
	var wg sync.WaitGroup
	go func() {
		defer wg.Done()
		<-ctx.Done()
		connection.Close()
	}()
	// send messages to server from stdin
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		writeToServer(ctx, connection)
	}()
	// read messages from server
	go func() {
		defer wg.Done()
		readFromServer(connection, cancel)
	}()
	wg.Add(3)

	wg.Wait()
}

func readFromServer(connection net.Conn, cancel context.CancelFunc) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			log.Println("Error reading:", err)
			cancel()
			break
		}
		log.Println("Received:", string(buffer[:mLen]))
	}
	log.Println("Stopped reading messages from server")
}

func writeToServer(ctx context.Context, connection net.Conn) {
	defer func() {
		log.Println("Stopped sending stdin inputs to server")
	}()

	inputs := make(chan string, 1)
	go func() {
		for {
			input := common.ScanFromStdin()
			inputs <- input
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case input := <-inputs:
			connection.Write([]byte(input))
		}
	}
}
