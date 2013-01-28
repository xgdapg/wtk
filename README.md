## Xgo
=======
Xgo是一个以简化web开发为目的的Go语言web框架。  
Xgo is a simple web framework to build webapp easily in Go.  

## Installation
============

    go get github.com/xgdapg/xgo

## Example
============
```go
package main

import (
	"github.com/xgdapg/xgo"
	"strconv"
	"strings"
)

func main() {
	xgo.RegisterController("/", &IndexController{})
	// /post/id123 与 /post/id123-2 会被路由到同一个控制器进行处理。
	// Both /post/id123 and /post/id123-2 will be routed to the same controller.
	xgo.RegisterController("/post/:id([0-9a-zA-Z]+)", &PostController{})
	xgo.RegisterController("/post/:id([0-9a-zA-Z]+)-:page([0-9]+)", &PostController{})
	// 注册一个钩子，当模板解析完成时会回调钩子函数进行处理。
	// Register a hook, and while the template has been parsed, the hook will be called.
	xgo.RegisterControllerHook(xgo.HookEventAfterRender, func(c *xgo.HookController) {
		if strings.HasPrefix(c.Context.Request.URL.Path, "/post") {
			c.Template.SetResultString(c.Template.GetResultString() + "<div>append a footer</div>")
		}
	})
	// 注册自定义404显示页面
	// Register a custom 404 page
	xgo.RegisterCustomHttpStatus(404, "notfound.html")
	xgo.Run()
}

type IndexController struct {
	xgo.Controller
}

func (this *IndexController) Get() {
	this.Context.WriteString("Hello, index page")
}

type PostController struct {
	xgo.Controller
}

func (this *PostController) Get() {
	id := this.Context.GetParam(":id")
	strPage := this.Context.GetParam(":page")
	page := 0
	if strPage != "" {
		page, _ = strconv.Atoi(strPage)
	}
	this.Template.SetVar("Title", "The post title")
	this.Template.SetVar("Content", "The post content")
	this.Template.SetVar("Id", id)
	this.Template.SetVar("Page", page)
	this.Template.SetTemplateFile("post.tpl")
}
```