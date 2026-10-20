package support

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/efegreat/forgo/forgotest"
	fhttp "github.com/efegreat/forgo/server/http"
)

// Client sends JSON requests to the HTTP server of an env.
type Client struct {
	router http.Handler // in process when set
	base   string       // otherwise, over loopback
}

// Router returns a Client that serves requests through the router of env's HTTP server, in process: the
// whole chain of the route runs, without the network. Feature tests use it.
func Router(t testing.TB, env *forgotest.Env) Client {
	t.Helper()
	h, err := env.Handler(fhttp.ServerNameHTTP)
	if err != nil {
		t.Fatal(err)
	}
	return Client{router: h}
}

// Live returns a Client that sends requests to env's HTTP server over loopback, as clients do. E2E tests
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
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return c.send(t, http.MethodPost, path, payload)
}

// Get requests path.
func (c Client) Get(t testing.TB, path string) *Response {
	t.Helper()
	return c.send(t, http.MethodGet, path, nil)
}

func (c Client) send(t testing.TB, method, path string, body []byte) *Response {
	t.Helper()
	var resp *http.Response
	if c.router != nil {
		req := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c.router.ServeHTTP(rec, req)
		resp = rec.Result()
	} else {
		req, err := http.NewRequestWithContext(t.Context(), method, c.base+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if resp, err = http.DefaultClient.Do(req); err != nil {
			t.Fatal(err)
		}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	r := &Response{Code: resp.StatusCode, body: raw}
	if err := json.Unmarshal(raw, &r.envelope); err != nil {
		t.Fatalf("%s %s: %d %s: %v", method, path, resp.StatusCode, raw, err)
	}
	return r
}

// Response is a response of the HTTP server, whose JSON envelope holds the result or the error.
type Response struct {
	Code     int
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
