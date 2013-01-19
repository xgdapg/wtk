package xgo

import (
	"mime"
	"net/http"
	"strings"
	"time"
)

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func (this *Context) WriteString(content string) {
	this.Response.Write([]byte(content))
}

func (this *Context) Abort(status int, content string) {
	this.Response.WriteHeader(status)
	this.WriteString(content)
}

func (this *Context) RedirectUrl(status int, url string) {
	this.SetHeader("Location", url)
	this.Response.WriteHeader(status)
}

func (this *Context) Redirect(url string) {
	this.RedirectUrl(302, url)
}

func (this *Context) NotModified() {
	this.Response.WriteHeader(304)
}

func (this *Context) NotFound(content string) {
	this.Response.WriteHeader(404)
	this.WriteString(content)
}

func (this *Context) SetHeader(name string, value string) {
	this.Response.Header().Set(name, value)
}

func (this *Context) AddHeader(name string, value string) {
	this.Response.Header().Add(name, value)
}

//Sets the content type by extension, as defined in the mime package. 
//For example, Context.ContentType("json") sets the content-type to "application/json"
func (this *Context) SetContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		this.SetHeader("Content-Type", ctype)
	}
}

//Sets a cookie -- duration is the amount of time in seconds. 0 = browser
func (this *Context) SetCookie(name string, value string, expires int64) {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}
	if expires > 0 {
		d := time.Duration(expires) * time.Second
		cookie.Expires = time.Now().Add(d)
	}
	http.SetCookie(this.Response, cookie)
}

func (this *Context) GetParam(name string) string {
	return this.Request.Form.Get(name)
}
