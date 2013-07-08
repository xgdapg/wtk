package wtk

import (
	"net/http"
)

type HandlerInterface interface {
	init(server *Server, w *wtkResponseWriter, r *http.Request)
	getHookHandler() *HookHandler
	context() *Context
	template() *Template
	session() *Session
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Render()
	Output()
}

type Handler struct {
	server   *Server
	Context  *Context
	Template *Template
	Session  *Session
}

func (this *Handler) init(server *Server, w *wtkResponseWriter, r *http.Request) {
	this.server = server

	this.Context = &Context{
		hdlr:           this,
		response:       w,
		ResponseWriter: w,
		Request:        r,
		pathVars:       nil,
		queryVars:      nil,
		formVars:       nil,
	}

	this.Template = &Template{
		hdlr:      this,
		tpl:       nil,
		vars:      nil,
		tplResult: nil,
	}

	this.Session = &Session{
		hdlr:           this,
		sessionManager: server.session,
		sessionId:      "",
		data:           nil,
		inited:         false,
	}
}

func (this *Handler) context() *Context {
	return this.Context
}

func (this *Handler) template() *Template {
	return this.Template
}

func (this *Handler) session() *Session {
	return this.Session
}

func (this *Handler) Get() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Post() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Delete() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Put() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Head() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Patch() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Options() {
	http.Error(this.Context.response, "Method Not Allowed", 405)
}

func (this *Handler) Render() {
	this.Template.Parse()
}

func (this *Handler) Output() {
	content := this.Template.GetResult()
	if len(content) > 0 {
		this.Context.WriteBytes(content)
	}
}

func (this *Handler) getHookHandler() *HookHandler {
	return &HookHandler{
		Context:  this.Context,
		Template: this.Template,
		Session:  this.Session,
	}
}
