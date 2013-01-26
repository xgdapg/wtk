package xgo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type xgoApp struct {
	Router  *xgoRouter
	Hook    *xgoHook
	Session *xgoSessionManager
}

func (this *xgoApp) init() *xgoApp {
	this.Router = &xgoRouter{
		app:         this,
		Rules:       []*xgoRoutingRule{},
		StaticRules: []*xgoRoutingRule{},
		StaticDir:   make(map[string]string),
	}
	this.Hook = &xgoHook{app: this}
	this.Session = new(xgoSessionManager)
	this.Session.RegisterStorage(new(xgoDefaultSessionStorage))
	return this
}

func (this *xgoApp) RegisterController(pattern string, c xgoControllerInterface) {
	this.Router.AddRule(pattern, c)
}

func (this *xgoApp) RegisterControllerHook(event string, hookFunc HookControllerFunc) {
	this.Hook.AddControllerHook(event, hookFunc)
}

func (this *xgoApp) SetStaticPath(sPath, fPath string) {
	this.Router.SetStaticPath(sPath, fPath)
}

func (this *xgoApp) RegisterSessionStorage(storage SessionStorageInterface) {
	this.Session.RegisterStorage(storage)
}

func (this *xgoApp) Run(mode string, addr string, port int) {
	listenAddr := net.JoinHostPort(addr, fmt.Sprintf("%d", port))
	var err error
	switch mode {
	case "http":
		err = http.ListenAndServe(listenAddr, this.Router)
	case "fcgi":
		l, e := net.Listen("tcp", listenAddr)
		if e != nil {
			panic("Fcgi listen error: " + e.Error())
		}
		err = fcgi.Serve(l, this.Router)
	default:
		err = http.ListenAndServe(listenAddr, this.Router)
	}
	if err != nil {
		panic("ListenAndServe error: " + err.Error())
	}
}

func (this *xgoApp) AppPath() string {
	path, _ := os.Getwd()
	return path
}
