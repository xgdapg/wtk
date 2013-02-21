package xgo

import (
	"net/http"
)

type xgoHandlerInterface interface {
	Init(*xgoApp, *xgoContext, *xgoTemplate, *xgoSession, string)
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
	app      *xgoApp
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

func (this *Handler) Init(app *xgoApp, ctx *xgoContext, tpl *xgoTemplate, sess *xgoSession, cn string) {
	this.app = app
	this.Context = ctx
	this.Context.hdlr = this
	this.Template = tpl
	this.Template.hdlr = this
	this.Session = sess
	this.Session.hdlr = this
}

func (this *Handler) Get() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Post() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Delete() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Put() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Head() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Patch() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Handler) Options() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
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
