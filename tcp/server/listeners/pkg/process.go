package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"owlsintheoven/learning-go/common"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Name       string
	Connection net.Conn
}

type Chat struct {
	From           string
	FromConnection net.Conn
	To             string
	Message        string
}

func ServeWithContext(ctx context.Context, serverType, host, port string, handleFunc func(context.Context, net.Conn)) {
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

	for {
		log.Println("Waiting for a client")
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			return
		}
		log.Println("Client connected")
		handleFunc(ctx, connection)
	}
}

func ServeMultipleClients(ctx context.Context, serverType, host, port string, handleFunc func(context.Context, net.Conn)) {
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
		log.Println("Waiting for a client")
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			break
		}
		log.Println("Client connected")
		wg.Add(1)
		go func() {
			handleFunc(ctx, connection)
			wg.Done()
		}()
	}
	wg.Wait()
}

func ServeMultipleClientsAndBroadcast(ctx context.Context, serverType, host, port string, handleFunc func(context.Context, net.Conn)) {
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
	clients := make(chan Client, 1)
	go func() {
		defer wg.Done()
		broadcast(ctx, clients)
	}()
	wg.Add(1)

	count := 1
	for {
		log.Println("Waiting for a client")
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			break
		}
		clients <- Client{
			Name:       strconv.Itoa(count),
			Connection: connection,
		}
		log.Println("Client connected")

		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			handleFunc(ctx, connection)
			clients <- Client{Name: strconv.Itoa(index)}
		}(count)
		count++
	}
	wg.Wait()
}

func ServeChatGroupWithContext(ctx context.Context, serverType, host, port string) {
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

	clients := make(chan Client, 1024)
	chats := make(chan Chat, 1024)

	var wg sync.WaitGroup
	go func() {
		defer wg.Done()
		chatgroup(ctx, clients, chats)
	}()
	wg.Add(1)

	for {
		log.Println("Waiting for a client")
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			break
		}
		log.Println("Client connected")
		wg.Add(1)
		go func() {
			defer wg.Done()
			name, err := askName(ctx, connection)
			if err != nil {
				return
			}
			clients <- Client{
				Name:       name,
				Connection: connection,
			}
			chat(ctx, connection, name, chats)
			clients <- Client{Name: name}
		}()
	}
	wg.Wait()
}

func chat(ctx context.Context, connection net.Conn, from string, chats chan Chat) {
	defer func() {
		log.Println("Stopped processing for client's chat", from)
	}()

	inputs := make(chan string, 1)
	go func() {
		for {
			msg, err := readFromClient(connection)
			inputs <- msg
			if err != nil {
				return
			}
		}
	}()
	for {
		for {
			select {
			case <-ctx.Done():
				connection.Close()
				return
			case input := <-inputs:
				if len(input) > 0 {
					splits := strings.Split(input, ":")
					if len(splits) != 2 {
						connection.Write([]byte("Unable to send a message in the chat group\nUsage: TO:MSG\nE.g. God:Hello World!"))
						continue
					}

					to, msg := splits[0], splits[1]
					chats <- Chat{
						From:           from,
						FromConnection: connection,
						To:             to,
						Message:        msg,
					}
				} else {
					return
				}
			}
		}
	}
}

func askName(ctx context.Context, connection net.Conn) (string, error) {
	go func() {
		<-ctx.Done()
		connection.Close()
	}()
	connection.Write([]byte("What's your name?\n"))
	return readFromClient(connection)
}

func chatgroup(ctx context.Context, clients chan Client, chats chan Chat) {
	defer func() {
		log.Println("Stopped updating all members")
	}()

	connections := make(map[string]net.Conn)

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-clients:
			var msg string
			if client.Connection != nil {
				connections[client.Name] = client.Connection
				msg = fmt.Sprintln(client.Name, "has joined the chat group")

				var names string
				for name, _ := range connections {
					names += fmt.Sprintf(", %s", name)
				}
				client.Connection.Write([]byte(fmt.Sprintf("Hi %s!\n", client.Name)))
				if len(names) != 0 {
					client.Connection.Write([]byte(fmt.Sprintf("There are %s in the chat group\n", names[2:])))
				}
			} else {
				delete(connections, client.Name)
				msg = fmt.Sprintln(client.Name, "has left the chat group")
			}
			log.Println(msg)
			for name, connection := range connections {
				if name != client.Name {
					connection.Write([]byte(msg))
				}
			}
		case chat := <-chats:
			log.Println(chat.From, "wants to tell", chat.To, ":", chat.Message)
			sent := false
			for name, connection := range connections {
				if name == chat.To {
					connection.Write([]byte(fmt.Sprintln(chat.From, "said:", chat.Message)))
					sent = true
					break
				}
			}
			if !sent {
				chat.FromConnection.Write([]byte(fmt.Sprintln(chat.To, "is not found in the chat group")))
			}
		}
	}
}

func Echo(ctx context.Context, connection net.Conn) {
	inputs := make(chan string, 1)
	go func() {
		for {
			msg, err := readFromClient(connection)
			inputs <- msg
			if err != nil {
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			connection.Close()
			return
		case input := <-inputs:
			if len(input) > 0 {
				connection.Write([]byte("Thanks! Got your message:" + input))
			} else {
				return
			}
		}
	}
}

func EchoAndWait(ctx context.Context, connection net.Conn) {
	inputs := make(chan string, 1)
	go func() {
		for {
			msg, err := readFromClient(connection)
			inputs <- msg
			if err != nil {
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			connection.Close()
			return
		case <-time.After(5 * time.Second):
			log.Println("Server has waited for client's input for too long")
			connection.Write([]byte("Zzzz\n"))
		case input := <-inputs:
			if len(input) > 0 {
				connection.Write([]byte("Thanks! Got your message:" + input))
			} else {
				return
			}
		}
	}
}

func readFromClient(connection net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Client has closed the connection")
		} else {
			log.Println("Error reading:", err.Error())
		}
		return "", err
	}
	log.Println("Received:", string(buffer[:mLen]))
	return string(buffer[:mLen]), nil
}

func broadcast(ctx context.Context, clients chan Client) {
	defer func() {
		log.Println("Stopped broadcasting")
	}()

	inputs := make(chan string, 1)
	go func() {
		for {
			input := common.ScanFromStdin()
			inputs <- input
		}
	}()
	connections := make(map[string]net.Conn)
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-clients:
			if client.Connection != nil {
				connections[client.Name] = client.Connection
			} else {
				delete(connections, client.Name)
			}
		case input := <-inputs:
			log.Println("connection count:", len(connections))
			for _, connection := range connections {
				connection.Write([]byte(input))
			}
		}
	}
}
