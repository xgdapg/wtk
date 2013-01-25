package xgo

import (
	"bytes"
	"compress/gzip"
	"mime"
	"net/http"
	"strings"
	"time"
)

type xgoContext struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func (this *xgoContext) WriteString(content string) {
	this.WriteBytes([]byte(content))
}

func (this *xgoContext) WriteBytes(content []byte) {
	this.SetHeader("Content-Type", http.DetectContentType(content))
	if EnableGzip {
		if strings.Contains(this.Request.Header.Get("Accept-Encoding"), "gzip") {
			this.SetHeader("Content-Encoding", "gzip")
			buf := new(bytes.Buffer)
			gz := gzip.NewWriter(buf)
			gz.Write(content)
			gz.Close()
			content = buf.Bytes()
		}
	}
	this.Response.Write(content)
}

func (this *xgoContext) Abort(status int, content string) {
	this.Response.WriteHeader(status)
	this.WriteString(content)
}

func (this *xgoContext) Redirect(status int, url string) {
	this.SetHeader("Location", url)
	this.Response.WriteHeader(status)
}

func (this *xgoContext) RedirectUrl(url string) {
	this.Redirect(302, url)
}

func (this *xgoContext) NotModified() {
	this.Response.WriteHeader(304)
}

func (this *xgoContext) NotFound(content string) {
	this.Response.WriteHeader(404)
	this.WriteString(content)
}

func (this *xgoContext) SetHeader(name string, value string) {
	this.Response.Header().Set(name, value)
}

func (this *xgoContext) AddHeader(name string, value string) {
	this.Response.Header().Add(name, value)
}

//Sets the content type by extension, as defined in the mime package. 
//For example, xgoContext.ContentType("json") sets the content-type to "application/json"
func (this *xgoContext) SetContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		this.SetHeader("Content-Type", ctype)
	}
}

//Sets a cookie -- duration is the amount of time in seconds. 0 = browser
func (this *xgoContext) SetCookie(name string, value string, expires int64) {
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

func (this *xgoContext) GetCookie(name string) string {
	cookie, err := this.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (this *xgoContext) GetParam(name string) string {
	return this.Request.Form.Get(name)
}
