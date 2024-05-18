package ggin

import (
	"log"
	"net/http"
)

var (
	default404Body = []byte("404 page not found")
)

type HandlerFunc func(*Context)

type Engine struct {
	routes map[string]map[string]HandlerFunc
}

func New() *Engine {
	engine := &Engine{
		routes: make(map[string]map[string]HandlerFunc),
	}
	return engine
}

func (engine *Engine) Handler() http.Handler {
	return engine
}

func (engine *Engine) Run(addr ...string) (err error) {
	defer func() { log.Println(err) }()

	address := resolveAddress(addr)
	log.Printf("Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, engine.Handler())
	return
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	c.Request = req
	c.Writer = w

	engine.handleHTTPRequest(c)
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path

	paths, ok := engine.routes[httpMethod]
	if ok {
		handler, ok := paths[rPath]
		if ok {
			handler(c)
			return
		}
	}
	serveError(c, default404Body)
}

func serveError(c *Context, defaultMessage []byte) {
	_, err := c.Writer.Write(defaultMessage)
	if err != nil {
		log.Printf("cannot write message to writer during serve error: %v\n", err)
	}
}

func (engine *Engine) handle(method string, path string, handler HandlerFunc) {
	paths, ok := engine.routes[method]
	if ok {
		paths[path] = handler
		engine.routes[method] = paths
	} else {
		router := map[string]HandlerFunc{
			path: handler,
		}
		engine.routes[method] = router
	}
}

func (engine *Engine) GET(path string, handler HandlerFunc) {
	engine.handle("GET", path, handler)
}

func (engine *Engine) POST(path string, handler HandlerFunc) {
	engine.handle("POST", path, handler)
}
