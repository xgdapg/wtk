package xgo

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"sync"
)

type App struct {
	Id               int
	listener         net.Listener
	router           *xgoRouter
	hook             *xgoHook
	session          *xgoSessionManager
	customHttpStatus map[int]string
	// extHook *xgoHook
}

func (this *App) init(id int) *App {
	this.Id = id
	this.router = &xgoRouter{
		app:            this,
		Routes:         []*Route{},
		StaticRoutes:   make(map[string]*Route),
		StaticFileDir:  make(map[string]int),
		StaticFileType: make(map[string]int),
		lock:           new(sync.Mutex),
		routeCache:     make(map[string]*xgoRouteCache),
	}
	this.hook = &xgoHook{app: this}
	// this.extHook = &xgoHook{app: this}
	this.session = new(xgoSessionManager)
	this.session.RegisterStorage(new(xgoDefaultSessionStorage))
	this.customHttpStatus = make(map[int]string)
	return this
}

func (this *App) AddRoute(pattern string, c HandlerInterface) *Route {
	return this.router.AddRoute(pattern, c)
}

func (this *App) RemoveRoute(pattern string) {
	this.router.RemoveRoute(pattern)
}

func (this *App) SetPrefixPath(prefix string) {
	this.router.SetPrefixPath(prefix)
}

func (this *App) AddHandlerHook(event string, hookFunc HookHandlerFunc) {
	this.hook.AddHandlerHook(event, hookFunc)
}

func (this *App) callHandlerHook(event string, hc *HookHandler) {
	this.hook.CallHandlerHook(event, hc)
	// this.extHook.CallHandlerHook(event, hc)
}

// func (this *App) registerAddonHandlerHook(event string, hookFunc HookHandlerFunc) {
// 	this.extHook.AddHandlerHook(event, hookFunc)
// }

// func (this *App) clearExtHook(event string, hookFunc HookHandlerFunc) {
// 	this.extHook = &xgoHook{app: this}
// }

func (this *App) AddStaticFileDir(dirs ...string) {
	this.router.AddStaticFileDir(dirs...)
}

func (this *App) RemoveStaticFileDir(dirs ...string) {
	this.router.RemoveStaticFileDir(dirs...)
}

func (this *App) AddStaticFileType(ext ...string) {
	this.router.AddStaticFileType(ext...)
}

func (this *App) RemoveStaticFileType(ext ...string) {
	this.router.RemoveStaticFileType(ext...)
}

func (this *App) RegisterSessionStorage(storage SessionStorageInterface) {
	this.session.RegisterStorage(storage)
}

func (this *App) RegisterCustomHttpStatus(code int, filePath string) {
	this.customHttpStatus[code] = filePath
}

func (this *App) Run(mode string, addr string, port int) error {
	listenAddr := net.JoinHostPort(addr, fmt.Sprintf("%d", port))

	var tlsConfig *tls.Config
	if mode == "https" {
		tlsConfig = &tls.Config{}
		if tlsConfig.NextProtos == nil {
			tlsConfig.NextProtos = []string{"http/1.1"}
		}
		var err error
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(SslCertificate, SslCertificateKey)
		if err != nil {
			return err
		}
	}

	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		return e
	}
	this.listener = l

	switch mode {
	case "http":
		http.Serve(l, this.router)
	case "fcgi":
		fcgi.Serve(l, this.router)
	case "https":
		http.Serve(tls.NewListener(l, tlsConfig), this.router)
	default:
		http.Serve(l, this.router)
	}
	l.Close()
	return nil
}

func (this *App) Stop() {
	this.listener.Close()
}

func (this *App) Close() {
	delete(apps, this.Id)
	this.Stop()
}

func (this *App) Clone() *App {
	a := NewApp()
	a.router = this.router
	a.hook = this.hook
	a.session = this.session
	a.customHttpStatus = this.customHttpStatus
	return a
}
