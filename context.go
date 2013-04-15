package xgo

import (
	"bufio"
	"bytes"
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

type Context struct {
	hdlr           *Handler
	response       *xgoResponseWriter
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (this *Context) finish() {
	this.response.Finished = true
	this.response.Close()
}

func (this *Context) WriteString(content string) {
	this.WriteBytes([]byte(content))
}

func (this *Context) WriteBytes(content []byte) {
	if this.response.Closed {
		return
	}
	hc := this.hdlr.getHookHandler()
	this.hdlr.app.callHandlerHook("BeforeOutput", hc)
	if this.response.Finished {
		return
	}

	this.SetHeader("Content-Type", http.DetectContentType(content))

	this.response.Write(content)

	this.hdlr.app.callHandlerHook("AfterOutput", hc)
	if this.response.Finished {
		return
	}
	this.response.Close()
}

func (this *Context) Abort(status int, content string) {
	this.response.WriteHeader(status)
	this.WriteString(content)
	this.finish()
}

func (this *Context) Redirect(status int, url string) {
	prefix := this.hdlr.app.router.PrefixPath
	http.Redirect(this.response, this.Request, prefix+url, status)
	this.finish()
}

func (this *Context) RedirectUrl(url string) {
	this.Redirect(302, url)
}

func (this *Context) NotModified() {
	this.response.WriteHeader(304)
	this.finish()
}

func (this *Context) NotFound() {
	this.response.WriteHeader(404)
	this.finish()
}

func (this *Context) SetHeader(name string, value string) {
	this.response.Header().Set(name, value)
}

func (this *Context) AddHeader(name string, value string) {
	this.response.Header().Add(name, value)
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
	http.SetCookie(this.response, cookie)
}

func (this *Context) GetCookie(name string) string {
	cookie, err := this.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (this *Context) SetSecureCookie(name string, value string, expires int64) {
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
	sig := util.getCookieSig(CookieSecret+this.Request.UserAgent()+strings.Split(this.Request.RemoteAddr, ":")[0], name, vs, ts)
	cookie := strings.Join([]string{vs, ts, sig}, "|")
	this.SetCookie(name, cookie, expires)
}

func (this *Context) GetSecureCookie(name string) string {
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
	if util.getCookieSig(CookieSecret+this.Request.UserAgent()+strings.Split(this.Request.RemoteAddr, ":")[0], name, val, timestamp) != sig {
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

func (this *Context) GetVar(name string) string {
	if this.Request.Form == nil {
		this.Request.ParseForm()
	}
	return this.Request.Form.Get(name)
}

func (this *Context) GetVars(name string) []string {
	if this.Request.Form == nil {
		this.Request.ParseForm()
	}
	vs, ok := this.Request.Form[name]
	if !ok || len(vs) == 0 {
		return []string{}
	}
	return vs
}

func (this *Context) GetUploadFile(name string) (*UploadFile, error) {
	if this.Request.Method != "POST" && this.Request.Method != "PUT" {
		return nil, errors.New("Incorrect method: " + this.Request.Method)
	}
	if this.Request.MultipartForm == nil {
		this.Request.ParseMultipartForm(0)
	}
	if this.Request.MultipartForm.File != nil {
		if fhs := this.Request.MultipartForm.File[name]; len(fhs) > 0 {
			uploadFile := &UploadFile{
				Filename:   fhs[0].Filename,
				fileHeader: fhs[0],
			}
			return uploadFile, nil
		}
	}
	return nil, http.ErrMissingFile
}

func (this *Context) GetUploadFiles(name string) ([]*UploadFile, error) {
	uploadFiles := []*UploadFile{}
	if this.Request.Method != "POST" && this.Request.Method != "PUT" {
		return uploadFiles, errors.New("Incorrect method: " + this.Request.Method)
	}
	if this.Request.MultipartForm == nil {
		this.Request.ParseMultipartForm(0)
	}
	if this.Request.MultipartForm.File != nil {
		if fhs := this.Request.MultipartForm.File[name]; len(fhs) > 0 {
			for _, fh := range fhs {
				uploadFiles = append(uploadFiles, &UploadFile{
					Filename:   fh.Filename,
					fileHeader: fh,
				})
			}
			return uploadFiles, nil
		}
	}
	return uploadFiles, http.ErrMissingFile
}

type UploadFile struct {
	Filename   string
	fileHeader *multipart.FileHeader
}

func (this *UploadFile) SaveFile(savePath string) error {
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

func (this *UploadFile) GetContentType() string {
	return this.fileHeader.Header.Get("Content-Type")
}

func (this *UploadFile) GetRawContentType() string {
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
