package wtk

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	hdlr           *Handler
	response       *wtkResponseWriter
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	pathVars       url.Values
	queryVars      url.Values
	formVars       url.Values
}

func (this *Context) GetPathVar(name string) string {
	if this.pathVars == nil {
		this.pathVars = make(url.Values)
	}
	return this.pathVars.Get(name)
}

func (this *Context) GetPathVars(name string) []string {
	if this.pathVars == nil {
		this.pathVars = make(url.Values)
	}
	vs, ok := this.pathVars[name]
	if !ok || len(vs) == 0 {
		return []string{}
	}
	return vs
}

func (this *Context) GetQueryVar(name string) string {
	if this.queryVars == nil {
		this.queryVars = this.Request.URL.Query()
	}
	return this.queryVars.Get(name)
}

func (this *Context) GetQueryVars(name string) []string {
	if this.queryVars == nil {
		this.queryVars = this.Request.URL.Query()
	}
	vs, ok := this.queryVars[name]
	if !ok || len(vs) == 0 {
		return []string{}
	}
	return vs
}

func (this *Context) parseForm() error {
	if this.formVars == nil {
		v := this.Request.Header.Get("Content-Type")
		d, _, err := mime.ParseMediaType(v)
		if err != nil {
			return err
		}
		if d == "multipart/form-data" {
			err := this.Request.ParseMultipartForm(0)
			if err != nil {
				return err
			}
			this.formVars = url.Values(this.Request.MultipartForm.Value)
		} else {
			err := this.Request.ParseForm()
			if err != nil {
				return err
			}
			this.formVars = this.Request.PostForm
		}
	}
	return nil
}

func (this *Context) GetFormVar(name string) string {
	this.parseForm()
	return this.formVars.Get(name)
}

func (this *Context) GetFormVars(name string) []string {
	this.parseForm()
	vs, ok := this.formVars[name]
	if !ok || len(vs) == 0 {
		return []string{}
	}
	return vs
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
	this.hdlr.callHandlerHook("BeforeOutput")
	if this.response.Finished {
		return
	}
	this.SetHeader("Content-Type", http.DetectContentType(content))
	if EnableGzip && len(content) < GzipMinLength {
		this.response.gzipWriter = nil
	}
	this.response.Write(content)

	this.hdlr.callHandlerHook("AfterOutput")
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
	if url == "" {
		return
	}
	if url[0] == '/' {
		prefix := this.hdlr.server.router.PrefixPath
		url = prefix + url
	}
	http.Redirect(this.response, this.Request, url, status)
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
func (this *Context) SetCookieWithArgs(name string, value string, maxage int, path string, domain string, secure bool, httponly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httponly,
	}
	if maxage > 0 {
		d := time.Duration(maxage) * time.Second
		cookie.Expires = time.Now().Add(d)
		cookie.MaxAge = maxage
	}
	if path == "" && this.hdlr.server.router.PrefixPath != "" {
		cookie.Path = this.hdlr.server.router.PrefixPath
	}
	http.SetCookie(this.response, cookie)
}

func (this *Context) SetCookie(name string, value string, maxage int) {
	this.SetCookieWithArgs(name, value, maxage, "", "", false, false)
}

func (this *Context) GetCookie(name string) string {
	cookie, err := this.Request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (this *Context) SetSecureCookieWithArgs(name string, value string, maxage int, path string, domain string, secure bool, httponly bool) {
	ts := "0"
	if maxage > 0 {
		d := time.Duration(maxage) * time.Second
		t := time.Now().Add(d)
		ts = strconv.FormatInt(t.Unix(), 10)
	}
	text := name + value + ts
	text += this.Request.UserAgent()

	sig := util.getCookieSig(CookieSecret, text)
	val := base64.URLEncoding.EncodeToString(util.AesEncrypt([]byte(CookieSecret), []byte(ts+"|"+value)))

	this.SetCookieWithArgs(name, sig+"|"+val, maxage, path, domain, secure, httponly)
}

func (this *Context) SetSecureCookie(name string, value string, maxage int) {
	this.SetSecureCookieWithArgs(name, value, maxage, "", "", false, true)
}

func (this *Context) GetSecureCookie(name string) string {
	str := this.GetCookie(name)
	if str == "" {
		return ""
	}
	strs := strings.SplitN(str, "|", 2)
	if len(strs) != 2 {
		return ""
	}
	sig := strs[0]
	val := strs[1]
	b, err := base64.URLEncoding.DecodeString(val)
	if err != nil {
		return ""
	}
	decrypted := string(util.AesDecrypt([]byte(CookieSecret), b))
	if decrypted == "" {
		return ""
	}
	parts := strings.SplitN(decrypted, "|", 2)
	if len(parts) != 2 {
		return ""
	}
	ts := parts[0]
	value := parts[1]

	text := name + value + ts
	text += this.Request.UserAgent()

	if util.getCookieSig(CookieSecret, text) != sig {
		return ""
	}
	expires, err := strconv.ParseInt(ts, 0, 64)
	if err != nil || expires > 0 && time.Now().Unix() > expires {
		return ""
	}
	return value
}

func (this *Context) GetUploadFile(name string) (*UploadFile, error) {
	if this.Request.Method != "POST" && this.Request.Method != "PUT" {
		return nil, errors.New("Incorrect method: " + this.Request.Method)
	}
	this.parseForm()
	if this.Request.MultipartForm == nil {
		return nil, http.ErrNotMultipart
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
	this.parseForm()
	if this.Request.MultipartForm == nil {
		return nil, http.ErrNotMultipart
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

func (this *UploadFile) SaveFile(savePath string) (int64, error) {
	file, err := this.fileHeader.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()
	f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, file)
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
