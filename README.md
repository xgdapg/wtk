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
)

func main() {
	xgo.RegisterController("/", &IndexController{})
	xgo.RegisterController("/post/:id([0-9a-zA-Z]+)", &PostController{})
	xgo.RegisterController("/post/:id([0-9a-zA-Z]+)-:page([0-9]+)", &PostController{})
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