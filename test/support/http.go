package support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/efegreat/forgo/forgotest"
	fhttp "github.com/efegreat/forgo/server/http"

	"example/config"
)

// Client sends JSON requests to an HTTP server of an env.
type Client struct {
	router http.Handler // in process when set
	remote string       // the client address of the requests in process
	base   string       // otherwise, over loopback
}

// clients numbers the clients of Router and Ops, so that each sends from an address of its own.
var clients atomic.Uint32

// Router returns a Client that serves requests through the router of env's API server, in process: the
// whole chain of the route runs, without the network. Feature tests use it.
//
// Each Client sends from an address of its own, as another client would, so that the rate limits of the
// API, which count per client, never carry over from one test to the next.
func Router(t testing.TB, env *forgotest.Env) Client {
	t.Helper()
	return inProcess(t, env, fhttp.ServerNameHTTP)
}

// Ops returns a Client of env's ops server, in process: its health, metrics and live config endpoints.
func Ops(t testing.TB, env *forgotest.Env) Client {
	t.Helper()
	return inProcess(t, env, config.Ops("", 0).Name())
}

func inProcess(t testing.TB, env *forgotest.Env, server string) Client {
	t.Helper()
	h, err := env.Handler(server)
	if err != nil {
		t.Fatal(err)
	}
	n := clients.Add(1)
	return Client{router: h, remote: fmt.Sprintf("198.18.%d.%d:40000", n>>8&0xff, n&0xff)}
}

// Live returns a Client that sends requests to env's API server over loopback, as clients do. E2E tests
// use it.
func Live(t testing.TB, env *forgotest.Env) Client {
	t.Helper()
	addr, err := env.Addr(fhttp.ServerNameHTTP)
	if err != nil {
		t.Fatal(err)
	}
	return Client{base: "http://" + addr}
}

// Post sends body, encoded as JSON, to path.
func (c Client) Post(t testing.TB, path string, body any) *Response {
	t.Helper()
	return c.send(t, http.MethodPost, path, encode(t, body), nil)
}

// Put sends body, encoded as JSON, to path, with header.
func (c Client) Put(t testing.TB, path string, body any, header http.Header) *Response {
	t.Helper()
	return c.send(t, http.MethodPut, path, encode(t, body), header)
}

// Get requests path.
func (c Client) Get(t testing.TB, path string) *Response {
	t.Helper()
	return c.send(t, http.MethodGet, path, nil, nil)
}

func encode(t testing.TB, body any) []byte {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func (c Client) send(t testing.TB, method, path string, body []byte, header http.Header) *Response {
	t.Helper()
	var resp *http.Response
	if c.router != nil {
		req := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
		req.RemoteAddr = c.remote
		setHeader(req, header)
		rec := httptest.NewRecorder()
		c.router.ServeHTTP(rec, req)
		resp = rec.Result()
	} else {
		req, err := http.NewRequestWithContext(t.Context(), method, c.base+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		setHeader(req, header)
		if resp, err = http.DefaultClient.Do(req); err != nil {
			t.Fatal(err)
		}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	r := &Response{Code: resp.StatusCode, Header: resp.Header, body: raw}
	if mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); mediaType == "application/json" {
		if err := json.Unmarshal(raw, &r.envelope); err != nil {
			t.Fatalf("%s %s: %d %s: %v", method, path, resp.StatusCode, raw, err)
		}
	}
	return r
}

func setHeader(req *http.Request, header http.Header) {
	req.Header.Set("Content-Type", "application/json")
	for name, values := range header {
		req.Header[name] = values
	}
}

// Response is a response of an HTTP server, whose JSON envelope holds the result or the error.
type Response struct {
	Code     int
	Header   http.Header
	body     []byte
	envelope struct {
		ErrCode string          `json:"errCode"`
		Result  json.RawMessage `json:"result"`
	}
}

// Status fails t unless the response has the status want.
func (r *Response) Status(t testing.TB, want int) {
	t.Helper()
	if r.Code != want {
		t.Fatalf("status %d, want %d: %s", r.Code, want, r.body)
	}
}

// Text returns the body of the response, such as the metrics of an ops server.
func (r *Response) Text() string {
	return string(r.body)
}

// HasSample reports whether the body of the response, metrics in Prometheus's text format, holds a sample of
// name with every label, such as `forgo_operation_name="place_order"`.
func (r *Response) HasSample(name string, labels ...string) bool {
	for line := range strings.SplitSeq(string(r.body), "\n") {
		if !strings.HasPrefix(line, name+"{") {
			continue
		}
		found := true
		for _, label := range labels {
			found = found && strings.Contains(line, label)
		}
		if found {
			return true
		}
	}
	return false
}

// ErrCode returns the error code of an error response, such as "validation".
func (r *Response) ErrCode() string {
	return r.envelope.ErrCode
}

// Result decodes the result of the response.
func Result[T any](t testing.TB, r *Response) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.envelope.Result, &v); err != nil {
		t.Fatalf("result %s: %v", r.envelope.Result, err)
	}
	return v
}
