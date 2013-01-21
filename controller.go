package xgo

import (
	"net/http"
)

type ControllerInterface interface {
	Init(*Context, *Template, string)
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
	Tpl *Template
}

func (this *Controller) Init(ctx *Context, tpl *Template, cn string) {
	this.Ctx = ctx
	this.Tpl = tpl
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
	this.Tpl.Parse()
}

func (this *Controller) Output() {
	this.Ctx.WriteString(this.Tpl.GetResult())
}

type ControllerHook struct {
	app   *App
	hooks []*controllerHookData
}

func (this *ControllerHook) AddHook(event string, hookFunc controllerHookFunc) {
	data := &controllerHookData{
		Event: event,
		Func:  hookFunc,
	}
	this.hooks = append(this.hooks, data)

}

func (this *ControllerHook) CallHook(event string, url string, hc *HookController) {
	for _, hook := range this.hooks {
		if hook.Event == event {
			hook.Func(url, hc)
			if hc.Ctx.Response.(*responseWriter).HasOutput {
				return
			}
		}
	}
}

type controllerHookFunc func(string, *HookController)

type controllerHookData struct {
	Event string
	Func  controllerHookFunc
}

type HookController struct {
	Ctx *Context
	Tpl *Template
}
