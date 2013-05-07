## WTK
WTK is a simple web toolkit to build webapp easily in Go.  

## Installation

    go get github.com/xgdapg/wtk

## GoDoc
[http://godoc.org/github.com/xgdapg/wtk](http://godoc.org/github.com/xgdapg/wtk)

## WIKI
[https://github.com/xgdapg/wtk/wiki](https://github.com/xgdapg/wtk/wiki)

## Example
```go
package main

import (
	"github.com/xgdapg/wtk"
	"strconv"
)

func main() {
	wtk.AddRoute("/", &IndexHandler{})
	// Both /post/id123 and /post/id123-2 will be routed to the same Handler.
	wtk.AddRoute("/post/{id}", &PageHandler{})
	wtk.AddRoute("/post/{id}-{page([0-9]+)}", &PageHandler{})
	wtk.Run()
}

type IndexHandler struct {
	wtk.Handler
}

func (this *IndexHandler) Get() {
	this.Context.WriteString("Hello, index page")
}

type PageHandler struct {
	wtk.Handler
}

func (this *PageHandler) Get() {
	id := this.Context.GetPathVar("id")
	strPage := this.Context.GetPathVar("page")
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

