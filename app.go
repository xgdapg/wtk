package xgo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type xgoApp struct {
	router *xgoRouter
	hook   *xgoControllerHook
}

func (this *xgoApp) init() *xgoApp {
	this.router = &xgoRouter{
		app:         this,
		Rules:       []*xgoRoutingRule{},
		StaticRules: []*xgoRoutingRule{},
		StaticDir:   make(map[string]string),
	}
	this.hook = &xgoControllerHook{
		app:   this,
		hooks: []*xgoControllerHookData{},
	}
	return this
}

func (this *xgoApp) RegisterController(pattern string, c xgoControllerInterface) {
	this.router.AddRule(pattern, c)
}

func (this *xgoApp) AddControllerHook(event string, hookFunc xgoControllerHookFunc) {
	this.hook.AddHook(event, hookFunc)
}

func (this *xgoApp) SetStaticPath(sPath, fPath string) {
	this.router.SetStaticPath(sPath, fPath)
}

func (this *xgoApp) Run(addr string, port int) {
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

func (this *xgoApp) AppPath() string {
	path, _ := os.Getwd()
	return path
}
