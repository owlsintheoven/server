package main

import (
	"owlsintheoven/learning-go/fileengine/handlers"
	"owlsintheoven/learning-go/ggin"
)

func newRouter() *ggin.Engine {
	r := ggin.New()
	r.GET("/ping", func(c *ggin.Context) {
		c.Writer.Write([]byte("pong"))
	})
	r.POST("/api/v1/filehasher", handlers.Filehash)
	return r
}
