package xgo

import (
	"net/http"
)

type HandlerInterface interface {
	Init(*App, *Context, *Template, *Session, string)
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
	app      *App
	Context  *Context
	Template *Template
	Session  *Session
}

func (this *Handler) Init(app *App, ctx *Context, tpl *Template, sess *Session, cn string) {
	this.app = app
	this.Context = ctx
	this.Context.hdlr = this
	this.Template = tpl
	this.Template.hdlr = this
	for n, v := range tplVars {
		tpl.SetVar(n, v)
	}
	this.Session = sess
	this.Session.hdlr = this
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
