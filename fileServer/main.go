package main

import (
	"log"
)

func main() {
	r := newRouter()
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("done")
}
