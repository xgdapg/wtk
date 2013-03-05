package xgo

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

type xgoApp struct {
	Id               int
	listener         net.Listener
	router           *xgoRouter
	hook             *xgoHook
	session          *xgoSessionManager
	customHttpStatus map[int]string
	// extHook *xgoHook
}

func (this *xgoApp) init(id int) *xgoApp {
	this.Id = id
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

func (this *xgoApp) RegisterHandler(pattern string, c xgoHandlerInterface) {
	this.router.AddRule(pattern, c)
}

func (this *xgoApp) RegisterHandlerHook(event string, hookFunc HookHandlerFunc) {
	this.hook.AddHandlerHook(event, hookFunc)
}

func (this *xgoApp) callHandlerHook(event string, hc *HookHandler) {
	this.hook.CallHandlerHook(event, hc)
	// this.extHook.CallHandlerHook(event, hc)
}

// func (this *xgoApp) registerAddonHandlerHook(event string, hookFunc HookHandlerFunc) {
// 	this.extHook.AddHandlerHook(event, hookFunc)
// }

// func (this *xgoApp) clearExtHook(event string, hookFunc HookHandlerFunc) {
// 	this.extHook = &xgoHook{app: this}
// }

func (this *xgoApp) SetStaticPath(sPath, fPath string) {
	this.router.SetStaticPath(sPath, fPath)
}

func (this *xgoApp) SetStaticFileType(ext string) {
	this.router.SetStaticFileType(ext)
}

func (this *xgoApp) RegisterSessionStorage(storage SessionStorageInterface) {
	this.session.RegisterStorage(storage)
}

func (this *xgoApp) RegisterCustomHttpStatus(code int, filePath string) {
	this.customHttpStatus[code] = filePath
}

func (this *xgoApp) Run(mode string, addr string, port int) error {
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

func (this *xgoApp) Stop() {
	this.listener.Close()
}

func (this *xgoApp) Close() {
	delete(apps, this.Id)
	this.Stop()
}

func (this *xgoApp) Clone() *xgoApp {
	a := NewApp()
	a.router = this.router
	a.hook = this.hook
	a.session = this.session
	a.customHttpStatus = this.customHttpStatus
	return a
}

func (this *xgoApp) AppPath() string {
	path, _ := os.Getwd()
	return path
}
