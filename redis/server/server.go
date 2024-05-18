package main

import (
	"context"
	"log"
	"net"
	"owlsintheoven/learning-go/redis/server/handlers"
	"owlsintheoven/learning-go/redis/server/redis_db"
	"sync"
)

type Server struct {
	db redis_db.DBInterface
}

func NewServer() *Server {
	return &Server{
		db: redis_db.NewDB(),
	}
}

func (s *Server) serve(ctx context.Context) {
	l, err := net.Listen("tcp", "localhost"+":"+"6379")
	if err != nil {
		log.Fatalln("Error listening:", err.Error())
		return
	}
	go func() {
		<-ctx.Done()
		l.Close()
	}()

	var wg sync.WaitGroup
	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("Error accepting:", err.Error())
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handle(ctx, c)
		}()
	}
	wg.Wait()
}

func (s *Server) handle(ctx context.Context, c net.Conn) {
	go func() {
		<-ctx.Done()
		c.Close()
	}()
	h := handlers.NewHandler(s.db, c)
	for {
		err := h.Process(ctx)
		if err != nil {
			break
		}
	}
}
