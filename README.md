## Xgo
Xgo是一个以简化web开发为目的的Go语言web框架。  
Xgo is a simple web framework to build webapp easily in Go.  

## Installation

    go get github.com/xgdapg/xgo

## Example
```go
package main

import (
	"github.com/xgdapg/xgo"
	"strconv"
	"strings"
)

func main() {
	xgo.RegisterHandler("/", &IndexHandler{})
	// /post/id123 与 /post/id123-2 会被路由到同一个控制器进行处理。
	// Both /post/id123 and /post/id123-2 will be routed to the same Handler.
	xgo.RegisterHandler("/post/:id([0-9a-zA-Z]+)", &PageHandler{})
	xgo.RegisterHandler("/post/:id([0-9a-zA-Z]+)-:page([0-9]+)", &PageHandler{})
	// 注册一个钩子，当模板解析完成时会回调钩子函数进行处理。
	// Register a hook, and while the template has been parsed, the hook will be called.
	xgo.RegisterHandlerHook(xgo.HookAfterRender, func(c *xgo.HookHandler) {
		if strings.HasPrefix(c.Context.Request.URL.Path, "/post") {
			c.Template.SetResultString(c.Template.GetResultString() + "<div>append a footer</div>")
		}
	})
	// 注册自定义404显示页面
	// Register a custom 404 page
	xgo.RegisterCustomHttpStatus(404, "notfound.html")
	xgo.Run()
}

type IndexHandler struct {
	xgo.Handler
}

func (this *IndexHandler) Get() {
	this.Context.WriteString("Hello, index page")
}

type PageHandler struct {
	xgo.Handler
}

func (this *PageHandler) Get() {
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

## Variables
#### ListenAddr
App listening address. (default: "")
#### ListenPort
App listening port. (default: 80) 
#### RunMode
Options: http, https, fcgi. (default: http)
#### CookieSecret
Secret key for secure cookie. (default: "foobar")  
Set it to a different string if you want to use secret cookie or session.
#### SessionName
The session id is stored in a cookie named with SessionName. (default: "XGOSESSID")
#### SessionTTL
The session live time in server side. Any operation with the session (get,set) will reset the time. (default: 900)
#### EnableGzip
Enable xgo to compress the response content with gzip. (default: true)  
If you are using fcgi mode behind a web server (like nginx) which is also using gzip, you may need to set EnableGzip to false.
#### GzipMinLength
(default: 1024)
#### GzipTypes
(default: text, javascript, css, xml)
#### SslCertificate

#### SslCertificateKey


## Hook
Xgo provides hook for us to control the request and response out of handler.  
For example, if we need a user authorization in each admin page, we can register a hook like this:

	xgo.RegisterHandlerHook(xgo.HookAfterInit, func(c *xgo.HookHandler) {
		if strings.HasPrefix(c.Context.Request.URL.Path, "/admin") {
			succ := checkUser()
			if !succ {
				c.Context.RedirectUrl("/admin/login")
			}
		}
	})

Currently, there are only handler hooks, and the hook events are:

	xgo.HookAfterInit          
	xgo.HookBeforeMethodGet    
	xgo.HookAfterMethodGet     
	xgo.HookBeforeMethodPost   
	xgo.HookAfterMethodPost    
	xgo.HookBeforeMethodHead   
	xgo.HookAfterMethodHead    
	xgo.HookBeforeMethodDelete 
	xgo.HookAfterMethodDelete  
	xgo.HookBeforeMethodPut    
	xgo.HookAfterMethodPut     
	xgo.HookBeforeMethodPatch  
	xgo.HookAfterMethodPatch   
	xgo.HookBeforeMethodOptions
	xgo.HookAfterMethodOptions 
	xgo.HookBeforeRender       
	xgo.HookAfterRender        
	xgo.HookBeforeOutput       
	xgo.HookAfterOutput        

## Handler
#### Context
  - Context.Response: http.ResponseWriter
  - Context.Request: *http.Request
  - Context.WriteString(content string)
  - Context.WriteBytes(content []byte)
  - Context.Abort(status int, content string)
  - Context.Redirect(status int, url string)
  - Context.RedirectUrl(url string)
  - Context.SetHeader(name string, value string)
  - Context.AddHeader(name string, value string)
  - Context.SetContentType(ext string)
  - Context.SetCookie(name string, value string, expires int64)
  - Context.GetCookie(name string) string
  - Context.SetSecureCookie(name string, value string, expires int64)
  - Context.GetSecureCookie(name string) string
  - Context.GetParam(name string) string
  - Context.GetUploadFile(name string) (*xgoUploadFile, error)

#### Template
A simple example:

	func (this *PageHandler) Get() {
		this.Template.SetVar("Title", "The post title")
		this.Template.SetVar("Content", "The post content")
		this.Template.SetTemplateFile("post.tpl")
	}
Xgo will automatic call the Template.Parse() if you didn't call it.  

use Template.SetSubTemplateFile if you need a sub-template inside the post.tpl:

	func (this *PageHandler) Get() {
		this.Template.SetVar("Title", "The post title")
		this.Template.SetVar("Content", "The post content")
		this.Template.SetTemplateFile("post.tpl")
		this.Template.SetSubTemplateFile("footer", "post_footer.tpl")
	}
Maybe you want to fetch the parsed content and save it as a static html file:

	func (this *PageHandler) Get() {
		this.Template.SetVar("Title", "The post title")
		this.Template.SetVar("Content", "The post content")
		this.Template.SetTemplateFile("post.tpl")
		this.Template.SetSubTemplateFile("footer", "post_footer.tpl")
		this.Template.Parse()
		content := this.Template.GetResultString()
		writeTheContentToSomewhere(content)
	}
The methods of Template:
  - xgo.AddTemplateFunc(name string, tplFunc interface{})
  - Template.SetVar(name string, value interface{})
  - Template.GetVar(name string) interface{}
  - Template.SetTemplateString(str string) bool
  - Template.SetTemplateFile(filename string) bool
  - Template.SetSubTemplateString(name, str string) bool
  - Template.SetSubTemplateFile(name, filename string) bool
  - Template.Parse() bool
  - Template.GetResult() []byte
  - Template.GetResultString() string
  - Template.SetResult(p []byte)
  - Template.SetResultString(s string)

#### Session
Sessions are stored in memory by default.  
To store them in database or other places, you need a new implementation of SessionStorageInterface and register it:

	xgo.RegisterSessionStorage(storage SessionStorageInterface)
Usage:

	func (this *PageHandler) Get() {
		this.Session.Set("name", "data")
		val := this.Session.Get("name")
		this.Session.Delete("name")
	}

#### Upload files
In xgo, there is an easy way to upload files.

	func (this *UploadHandler) Post() {
		f, err := this.Context.GetUploadFile("userfile")
		if err != nil {
			log.Println(err)
			this.Context.RedirectUrl("/")
		}
		err = f.SaveFile("upload/" + f.Filename)
		if err != nil {
			log.Println(err)
			this.Context.RedirectUrl("/")
		}
	}
The returned variable f has several members:
  - f.Filename: the filename of the uploaded file.
  - f.SaveFile(savePath): save the uploaded file to the savePath
  - f.GetContentType(): return the Content-Type of the uploaded file, detected with request header.
  - f.GetRawContentType(): return the Content-Type of the uploaded file, detected with http.DetectContentType().

## Config
You can set the values of all the variables above in a config file.  
By default, xgo reads "app.conf" as config file in the same folder with app, and you can run your app like "./app configFilePath" to let xgo read the config file from "configFilePath".  
The config file format is json:  

	{
		"ListenAddr": "",
		"ListenPort": 8080,
		"EnableGzip": true,
		"GzipTypes": ["text", "javascript", "css", "xml"],
		"CustomString": "string value",
		"CustomStringArray": ["string1", "string2", "string3"]
		...
	}
As you see, you can also add some custom keys to config file, and fetch them with

	// first define a config struct
	type customConfig struct {
		CustomString      string
		CustomStringArray []string
	}
	// set default value
	cfg := &customConfig{
		CustomString:      "default string",
		CustomStringArray: []string{"default string1", "default string2"},
	}
	// load the config into this struct
	xgo.LoadConfig(cfg)
	// now the cfg.CustomString is "string value", or "default string" if there's no "CustomString" field in the config file or some errors occurred
	fmt.Println(cfg.CustomString)
If there is a method "OnLoaded" in your config struct (eg. customConfig), it will be called after everytime you load/reload the config.

	func (this *customConfig) OnLoaded() {
		fmt.Println("cfg.CustomString is changed to ", this.CustomString)
	}

## Custom error pages
It is usually very useful to have a custom 404 page. In xgo, we can register a 404 page like this:

	xgo.RegisterCustomHttpStatus(404, "notfound.html")
or other http status code, for example, 403

	xgo.RegisterCustomHttpStatus(403, "forbidden.html")

## .
To be continued.