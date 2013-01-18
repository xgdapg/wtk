package xgo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
)

type App struct {
	router *Router
}

func (this *App) RegisterController(pattern string, c ControllerInterface) {
	this.router.AddRule(pattern, c)
}

func (this *App) Run(addr string, port int) {
	listenAddr := fmt.Sprintf("%s:%d", addr, port)
	var err error
	switch RunMode {
	case "http":
		err = http.ListenAndServe(listenAddr, this.router)
	case "fcgi":
		l, e := net.Listen("tcp", listenAddr)
		if e != nil {
			panic("Fcgi listen error: " + e.Error())
		}
		err = fcgi.Serve(l, this.router)
	default:
		err = http.ListenAndServe(listenAddr, this.router)
	}
	if err != nil {
		panic("ListenAndServe error: " + err.Error())
	}
}
