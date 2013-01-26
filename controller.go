package xgo

import (
	"net/http"
)

type xgoControllerInterface interface {
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

type Controller struct {
	app      *xgoApp
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

func (this *Controller) Init(app *xgoApp, ctx *xgoContext, tpl *xgoTemplate, sess *xgoSession, cn string) {
	this.app = app
	this.Context = ctx
	this.Context.ctlr = this
	this.Template = tpl
	this.Template.ctlr = this
	this.Session = sess
	this.Session.ctlr = this
}

func (this *Controller) Get() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Post() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Delete() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Put() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Head() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Patch() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Options() {
	http.Error(this.Context.Response, "Method Not Allowed", 405)
}

func (this *Controller) Render() {
	this.Template.Parse()
}

func (this *Controller) Output() {
	this.Context.WriteBytes(this.Template.GetResult())
}

func (this *Controller) getHookController() *HookController {
	return &HookController{
		Context:  this.Context,
		Template: this.Template,
		Session:  this.Session,
	}
}
