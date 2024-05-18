package listeners

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"owlsintheoven/learning-go/common"
	"owlsintheoven/learning-go/tcp/server/listeners/models"
	"sync"
	"syscall"
)

func SOCKS4(serverType string, host string, port string) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		common.WaitForShutdown(sigs)
		cancel()
	}()

	serve(ctx, serverType, host, port)
}

func serve(ctx context.Context, serverType, host, port string) {
	listener, err := net.Listen(serverType, host+":"+port)
	if err != nil {
		log.Fatalln("Error listening:", err.Error())
		return
	}
	log.Println("Listening on " + host + ":" + port)
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	var wg sync.WaitGroup

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			break
		}
		log.Println("Client connected")
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle(ctx, connection)
		}()
	}
	wg.Wait()
}

func handle(ctx context.Context, clientConnection net.Conn) {
	defer func() {
		clientConnection.Close()
	}()
	go func() {
		<-ctx.Done()
		clientConnection.Close()
	}()

	clientReader := bufio.NewReader(clientConnection)
	socks4Req, err := models.ParseSocks4Request(clientReader)
	if err != nil {
		log.Printf("error parsing socks4 request: %s\n", err.Error())
		clientConnection.Write(models.FormResponse(models.REQUEST_REJECTED_OR_FAILED))
		return
	}

	log.Printf("request: command %s, port %s, ip %s, userID %s\n", socks4Req.CMD, socks4Req.GetPortString(), socks4Req.GetIPString(), socks4Req.UserID)

	if socks4Req.IsConnect() {
		processConnect(ctx, clientConnection, socks4Req.GetIPString(), socks4Req.GetPortString())
	} else {
		log.Println("Sorry BIND mode is not ready")
	}
}

func processConnect(ctx context.Context, clientConnection net.Conn, dstIP, dstPort string) {
	serverConnection, err := net.Dial("tcp", dstIP+":"+dstPort)
	if err != nil {
		log.Println("Error connecting to server: ", err.Error())
		clientConnection.Write(models.FormResponse(models.REQUEST_REJECTED_OR_FAILED))
		return
	}

	clientConnection.Write(models.FormResponse(models.REQUEST_GRANTED))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		relayData(ctx, clientConnection, serverConnection, "client->server")
	}()
	go func() {
		defer wg.Done()
		relayData(ctx, serverConnection, clientConnection, "server->client")
	}()
	wg.Wait()
}

func relayData(ctx context.Context, fromConnection, toConnection net.Conn, name string) {
	defer func() {
		log.Printf("%s: Stopped relaying data\n", name)
	}()
	inputs := make(chan []byte, 1024)
	go func() {
		for {
			buffer := make([]byte, 1024)
			mLen, err := fromConnection.Read(buffer)
			if err != nil {
				log.Printf("%s: Error reading %s\n", name, err.Error())
				inputs <- nil
				return
			}
			//log.Println("local address:", fromConnection.LocalAddr().String())
			//log.Println("remote address:", fromConnection.RemoteAddr().String())
			log.Printf("%s: Received %d bytes\n", name, mLen)
			inputs <- buffer[:mLen]
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case input := <-inputs:
			if len(input) > 0 {
				_, err := toConnection.Write(input)
				if err != nil {
					log.Printf("%s error writing %s\n", name, err.Error())
					toConnection.Close()
					return
				}
			} else {
				toConnection.Close()
				return
			}
		}
	}
}
