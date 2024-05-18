package common

import (
	"fmt"
	"os"
)

func WaitForShutdown(sigs chan os.Signal) {
	sig := <-sigs
	fmt.Printf("Signal received: %s\n", sig)
}
