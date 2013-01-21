## Xgo
=======
Xgo is a simple web framework to build webapp easily in Go.  
It was modified based on [astaxie/beego](https://github.com/astaxie/beego) at first, and now it has changed a lot.  
For self-use. Unfinished yet :)

## Installation
============
To install:

    go get github.com/xgdapg/xgo

## Example
============
```go
package main

import (
	"github.com/xgdapg/xgo"
)

func main() {
	xgo.RegisterController("/", &MainController{})
	xgo.RegisterController("/:id([0-9a-zA-Z]+)-:page([0-9]+)", &MainController{})
	xgo.SetStaticPath("/static", "static")
	xgo.AddControllerHook("AfterInit", func(url string, c *xgo.HookController) {
		c.Tpl.SetVar("Subb", "Hook edit Subb, url:"+url)
	})
	xgo.Run()
}

type MainController struct {
	xgo.Controller
}

func (this *MainController) Get() {
	id := this.Ctx.GetParam(":id")
	this.Tpl.SetVar("Title", "The test title")
	this.Tpl.SetVar("Content", "The test content")
	this.Tpl.SetVar("ID", "Div data")
	// this.Tpl.SetVar("Subb", "Subb data")
	this.Tpl.SetTemplate("index.tpl")
	this.Tpl.SetSubTemplate("Sub", "div.tpl")
	this.Tpl.SetSubTemplate("Subsub", "sub.tpl")
}
```