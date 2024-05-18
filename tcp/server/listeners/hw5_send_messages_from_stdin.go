package listeners

import (
	"context"
	"os"
	"os/signal"
	"owlsintheoven/learning-go/common"
	"owlsintheoven/learning-go/tcp/server/listeners/pkg"
	"syscall"
)

func ProcessConnectionsAndBroadcast(serverType string, host string, port string) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		common.WaitForShutdown(sigs)
		cancel()
	}()

	pkg.ServeMultipleClientsAndBroadcast(ctx, serverType, host, port, pkg.Echo)
}
