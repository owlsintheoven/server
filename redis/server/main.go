package main

import (
	"context"
	"os"
	"owlsintheoven/learning-go/common"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	go func() {
		common.WaitForShutdown(sigs)
		cancel()
	}()

	server := NewServer()
	server.serve(ctx)
}
