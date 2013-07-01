package wtk

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var testServer *Server

func init() {
	testServer = NewServer()
	testServer.AddRoute("/post", &PostHandler{})
	testServer.AddRoute("/post/{name([a-zA-Z0-9]+)}-{page([0-9]+)}", &PostHandler{})
	testServer.AddRoute("/cookie", &CookieHandler{})
	testServer.AddRoute("/{key(.*)}", &IndexHandler{})
}

func request(method, path string, body map[string]string, cookie map[string]string) *httptest.ResponseRecorder {
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
	if cookie != nil {
		for k, v := range cookie {
			r.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
				Path:  "/",
			})
		}
	}
	w := httptest.NewRecorder()
	testServer.router.ServeHTTP(w, r)
	return w
}

type testRequest struct {
	Method         string
	Path           string
	Body           map[string]string
	Cookie         map[string]string
	ExpectedStatus int
	ExpectedBody   string
	ExpectedCookie map[string]string
}

var testData = []*testRequest{
	{"GET", "/",
		nil, nil,
		200, "Index_Get", nil},
	{"GET", "/asdf",
		nil, nil,
		200, "Index_Get_asdf", nil},
	{"GET", "/post",
		nil, nil,
		200, "Post_Get", nil},
	{"GET", "/post/asdf",
		nil, nil,
		404, "", nil},
	{"GET", "/post/asdf-1",
		nil, nil,
		200, "Post_Get_asdf_1", nil},
	{"POST", "/post/asdf-1",
		map[string]string{"postname": "fdsa"}, nil,
		200, "Post_Post_fdsa", nil},
	{"GET", "/cookie",
		nil, map[string]string{"cookiename": "cookievalue", "securename": "29b5ebdb3686d0250f44929764e9a20b2616558e|0Hlde1JxXYhTO8fOWw=="},
		200, "cookievalue,securevalue", map[string]string{"newname1": "newvalue1", "newname2": "newvalue2", "securename": "29b5ebdb3686d0250f44929764e9a20b2616558e|0Hlde1JxXYhTO8fOWw=="}},
}

func TestRequest(t *testing.T) {
	for _, row := range testData {
		response := request(row.Method, row.Path, row.Body, row.Cookie)
		code := response.Code
		body := response.Body.String()
		header := response.Header()

		if code != row.ExpectedStatus {
			t.Fatalf("Path %s want status %d, but got %d", row.Path, row.ExpectedStatus, code)
		}
		if body != row.ExpectedBody {
			t.Fatalf("Path %s want body '%s', but got '%s'", row.Path, row.ExpectedBody, body)
		}
		if row.ExpectedCookie != nil {
			cookies := readSetCookies(header)
			for k, v := range row.ExpectedCookie {
				if gv, ok := cookies[k]; ok {
					if gv != v {
						t.Fatalf("Path %s want cookie %s='%s', but got '%s'", row.Path, k, v, gv)
					}
				} else {
					t.Fatalf("Path %s want cookie %s='%s', but got nothing", row.Path, k, v)
				}
			}
		}
		t.Log("PASS:", row.Method, row.Path)
	}
}

func readSetCookies(h http.Header) map[string]string {
	result := make(map[string]string)
	for _, line := range h["Set-Cookie"] {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		parts[0] = strings.TrimSpace(parts[0])
		j := strings.Index(parts[0], "=")
		if j < 0 {
			continue
		}
		name, value := parts[0][:j], parts[0][j+1:]
		result[name] = value
	}
	return result
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

type CookieHandler struct {
	Handler
}

func (this *CookieHandler) Get() {
	cv := this.Context.GetCookie("cookiename")
	scv := this.Context.GetSecureCookie("securename")
	this.Context.SetCookie("newname1", "newvalue1", 0)
	this.Context.SetCookie("newname2", "newvalue2", 0)
	this.Context.SetSecureCookie("securename", "securevalue", 0)
	this.Context.WriteString(cv + "," + scv)
}
