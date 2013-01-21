package xgo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type App struct {
	router *Router
}

func (this *App) init() *App {
	this.router = &Router{
		app:         this,
		Rules:       []*RoutingRule{},
		StaticRules: []*RoutingRule{},
		StaticDir:   make(map[string]string),
	}
	return this
}

func (this *App) RegisterController(pattern string, c ControllerInterface) {
	this.router.AddRule(pattern, c)
}

func (this *App) SetStaticPath(sPath, fPath string) {
	this.router.SetStaticPath(sPath, fPath)
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

func (this *App) AppPath() string {
	path, _ := os.Getwd()
	return path
}
