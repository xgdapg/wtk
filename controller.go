package xgo

import (
	"net/http"
)

type xgoControllerInterface interface {
	Init(*xgoContext, *xgoTemplate, *xgoSession, string)
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
	Context  *xgoContext
	Template *xgoTemplate
	Session  *xgoSession
}

func (this *Controller) Init(ctx *xgoContext, tpl *xgoTemplate, sess *xgoSession, cn string) {
	this.Context = ctx
	this.Template = tpl
	this.Session = sess
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
	this.Context.WriteString(this.Template.GetResult())
}
