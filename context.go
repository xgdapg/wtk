package xgo

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type xgoContext struct {
	hdlr     *Handler
	Response *xgoResponseWriter
	Request  *http.Request
}

func (this *xgoContext) Finish() {
	this.Response.Finished = true
	this.Response.Close()
}

func (this *xgoContext) WriteString(content string) {
	this.WriteBytes([]byte(content))
}

func (this *xgoContext) WriteBytes(content []byte) {
	if this.Response.Closed {
		return
	}
	hc := this.hdlr.getHookHandler()
	this.hdlr.app.callHandlerHook("BeforeOutput", hc)
	if this.Response.Finished {
		return
	}

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

	this.hdlr.app.callHandlerHook("AfterOutput", hc)
	if this.Response.Finished {
		return
	}
	this.Response.Close()
}

func (this *xgoContext) Abort(status int, content string) {
	this.Response.WriteHeader(status)
	this.WriteString(content)
	this.Finish()
}

func (this *xgoContext) Redirect(status int, url string) {
	this.SetHeader("Location", url)
	this.Response.WriteHeader(status)
	this.Finish()
}

func (this *xgoContext) RedirectUrl(url string) {
	this.Redirect(302, url)
}

func (this *xgoContext) NotModified() {
	this.Response.WriteHeader(304)
	this.Finish()
}

func (this *xgoContext) NotFound() {
	this.Response.WriteHeader(404)
	this.Finish()
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

func (this *xgoContext) SetSecureCookie(name string, value string, expires int64) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(value))
	encoder.Close()
	vs := buf.String()
	ts := "0"
	if expires > 0 {
		d := time.Duration(expires) * time.Second
		ts = strconv.FormatInt(time.Now().Add(d).Unix(), 10)
	}

	sig := util.getCookieSig(CookieSecret, name, vs, ts)
	cookie := strings.Join([]string{vs, ts, sig}, "|")
	this.SetCookie(name, cookie, expires)
}

func (this *xgoContext) GetSecureCookie(name string) string {
	value := this.GetCookie(name)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, "|", 3)
	if len(parts) != 3 {
		return ""
	}
	val := parts[0]
	timestamp := parts[1]
	sig := parts[2]
	if util.getCookieSig(CookieSecret, name, val, timestamp) != sig {
		return ""
	}
	ts, _ := strconv.ParseInt(timestamp, 0, 64)
	if ts > 0 && time.Now().Unix() > ts {
		return ""
	}
	buf := bytes.NewBufferString(val)
	encoder := base64.NewDecoder(base64.StdEncoding, buf)
	res, _ := ioutil.ReadAll(encoder)
	return string(res)
}

func (this *xgoContext) GetParam(name string) string {
	return this.Request.Form.Get(name)
}

func (this *xgoContext) GetUploadFile(name string) (*xgoUploadFile, error) {
	if this.Request.Method != "POST" && this.Request.Method != "PUT" {
		return nil, errors.New("Incorrect method: " + this.Request.Method)
	}
	if this.Request.MultipartForm != nil && this.Request.MultipartForm.File != nil {
		if fhs := this.Request.MultipartForm.File[name]; len(fhs) > 0 {
			uploadFile := &xgoUploadFile{
				Filename:   fhs[0].Filename,
				fileHeader: fhs[0],
			}
			return uploadFile, nil
		}
	}
	return nil, http.ErrMissingFile
}

type xgoUploadFile struct {
	Filename   string
	fileHeader *multipart.FileHeader
}

func (this *xgoUploadFile) SaveFile(savePath string) error {
	file, err := this.fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	return nil
}

func (this *xgoUploadFile) GetContentType() string {
	return this.fileHeader.Header.Get("Content-Type")
}

func (this *xgoUploadFile) GetRawContentType() string {
	file, err := this.fileHeader.Open()
	if err != nil {
		return ""
	}
	defer file.Close()
	r := bufio.NewReader(file)
	p := []byte{}
	for i := 0; i < 512; i++ {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		p = append(p, b)
	}
	return http.DetectContentType(p)
}
