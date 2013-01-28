package xgo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type xgoApp struct {
	router           *xgoRouter
	hook             *xgoHook
	session          *xgoSessionManager
	customHttpStatus map[int]string
	// extHook *xgoHook
}

func (this *xgoApp) init() *xgoApp {
	this.router = &xgoRouter{
		app:         this,
		Rules:       []*xgoRoutingRule{},
		StaticRules: []*xgoRoutingRule{},
		StaticDir:   make(map[string]string),
	}
	this.hook = &xgoHook{app: this}
	// this.extHook = &xgoHook{app: this}
	this.session = new(xgoSessionManager)
	this.session.RegisterStorage(new(xgoDefaultSessionStorage))
	this.customHttpStatus = make(map[int]string)
	return this
}

func (this *xgoApp) RegisterController(pattern string, c xgoControllerInterface) {
	this.router.AddRule(pattern, c)
}

func (this *xgoApp) RegisterControllerHook(event string, hookFunc HookControllerFunc) {
	this.hook.AddControllerHook(event, hookFunc)
}

func (this *xgoApp) callControllerHook(event string, hc *HookController) {
	this.hook.CallControllerHook(event, hc)
	// this.extHook.CallControllerHook(event, hc)
}

// func (this *xgoApp) registerAddonControllerHook(event string, hookFunc HookControllerFunc) {
// 	this.extHook.AddControllerHook(event, hookFunc)
// }

// func (this *xgoApp) clearExtHook(event string, hookFunc HookControllerFunc) {
// 	this.extHook = &xgoHook{app: this}
// }

func (this *xgoApp) SetStaticPath(sPath, fPath string) {
	this.router.SetStaticPath(sPath, fPath)
}

func (this *xgoApp) RegisterSessionStorage(storage SessionStorageInterface) {
	this.session.RegisterStorage(storage)
}

func (this *xgoApp) RegisterCustomHttpStatus(code int, filePath string) {
	this.customHttpStatus[code] = filePath
}

func (this *xgoApp) Run(mode string, addr string, port int) {
	listenAddr := net.JoinHostPort(addr, fmt.Sprintf("%d", port))
	var err error
	switch mode {
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
