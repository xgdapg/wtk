package xgo

import (
	"net/http"
)

type ControllerInterface interface {
	Init(ct *Context, cn string)
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
	Ctx *Context
}

func (this *Controller) Init(ctx *Context, cn string) {
	this.Ctx = ctx
}

func (this *Controller) Get() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Post() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Delete() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Put() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Head() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Patch() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Options() {
	http.Error(this.Ctx.Response, "Method Not Allowed", 405)
}

func (this *Controller) Render() {
	// 
}

func (this *Controller) Output() {
	// 
}

func (this *Controller) SetVar(k string, v interface{}) {

}
