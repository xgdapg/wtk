package wtk

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

var testServer *Server

func init() {
	testServer = NewServer()
	testServer.AddRoute("/post", &PostHandler{})
	testServer.AddRoute("/post/{name([a-zA-Z0-9]+)}-{page([0-9]+)}", &PostHandler{})
	testServer.AddRoute("/{key(.*)}", &IndexHandler{})
}

type tResponseWriter struct {
	header http.Header
	body   *bytes.Buffer
	code   int
}

func (this *tResponseWriter) Write(p []uint8) (n int, err error) {
	return this.body.Write(p)
}

func (this *tResponseWriter) Header() http.Header {
	return this.header
}

func (this *tResponseWriter) WriteHeader(code int) {
	this.code = code
}

func request(method, path string, body map[string]string) (int, string) {
	var reqBody io.ReadCloser
	bodyType := ""
	if method == "POST" && body != nil {
		uv := make(url.Values)
		for k, v := range body {
			uv.Add(k, v)
		}
		reqBody = ioutil.NopCloser(bytes.NewBufferString(uv.Encode()))
		bodyType = "application/x-www-form-urlencoded"
	}
	r, _ := http.NewRequest(method, path, reqBody)
	if bodyType != "" {
		r.Header.Set("Content-Type", bodyType)
	}
	w := &tResponseWriter{
		header: make(http.Header),
		body:   new(bytes.Buffer),
		code:   200,
	}
	testServer.router.ServeHTTP(w, r)
	return w.code, w.body.String()
}

type testRequest struct {
	Method         string
	Path           string
	Body           map[string]string
	ExpectedStatus int
	ExpectedBody   string
}

var testData = []*testRequest{
	{"GET", "/", nil, 200, "Index_Get"},
	{"GET", "/asdf", nil, 200, "Index_Get_asdf"},
	{"GET", "/post", nil, 200, "Post_Get"},
	{"GET", "/post/asdf", nil, 404, ""},
	{"GET", "/post/asdf-1", nil, 200, "Post_Get_asdf_1"},
	{"POST", "/post/asdf-1", map[string]string{"postname": "fdsa"}, 200, "Post_Post_fdsa"},
}

func TestRequest(t *testing.T) {
	for _, row := range testData {
		code, body := request(row.Method, row.Path, row.Body)
		if code == row.ExpectedStatus && body == row.ExpectedBody {
			t.Log("pass:", row.Method, row.Path)
		} else {
			t.Fatalf("Path %s want status %d and body '%s', but got %d and '%s'", row.Path, row.ExpectedStatus, row.ExpectedBody, code, body)
		}
	}
}

type IndexHandler struct {
	Handler
}

func (this *IndexHandler) Get() {
	key := this.Context.GetPathVar("key")
	if key == "" {
		this.Context.WriteString("Index_Get")
	} else {
		this.Context.WriteString("Index_Get_" + key)
	}

}

type PostHandler struct {
	Handler
}

func (this *PostHandler) Get() {
	name := this.Context.GetPathVar("name")
	page := this.Context.GetPathVar("page")
	if name == "" && page == "" {
		this.Context.WriteString("Post_Get")
	} else {
		this.Context.WriteString("Post_Get_" + name + "_" + page)
	}

}

func (this *PostHandler) Post() {
	this.Context.WriteString("Post_Post_" + this.Context.GetFormVar("postname"))
}
