package portal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

type testServer struct {
	*httptest.Server
	form url.Values
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func newTestServer(t *testing.T, wantPath, fixture string) *testServer {
	t.Helper()
	ts := &testServer{form: url.Values{}}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want %q", got, "application/x-www-form-urlencoded")
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		ts.form = r.Form
		w.Write(readFixture(t, fixture))
	}))
	return ts
}

func wantForm(t *testing.T, form url.Values, want map[string]string) {
	t.Helper()
	for k, v := range want {
		if got := form.Get(k); got != v {
			t.Errorf("form[%s] = %q, want %q", k, got, v)
		}
	}
}

func wantEpochA(t *testing.T, form url.Values) {
	t.Helper()
	a := form.Get("a")
	ms, err := strconv.ParseInt(a, 10, 64)
	if err != nil || ms <= 0 {
		t.Errorf("form[a] = %q, want a positive epoch-millis integer", a)
	}
}

func TestLoginLive(t *testing.T) {
	ts := newTestServer(t, "/login.xml", "login_live.xml")
	defer ts.Close()
	c := New(ts.URL, ts.Client())

	ok, msg, err := c.Login(context.Background(), "1BI22CS123", "hunter2")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
	if msg != "authenticated" {
		t.Errorf("msg = %q, want %q", msg, "authenticated")
	}

	wantForm(t, ts.form, map[string]string{
		"mode":        "191",
		"username":    "1BI22CS123",
		"password":    "hunter2",
		"producttype": "0",
	})
	wantEpochA(t, ts.form)
}

func TestLoginFailure(t *testing.T) {
	ts := newTestServer(t, "/login.xml", "login_fail.xml")
	defer ts.Close()
	c := New(ts.URL, ts.Client())

	ok, msg, err := c.Login(context.Background(), "1BI22CS123", "wrongpass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if ok {
		t.Error("ok = true, want false")
	}
	if msg != "Invalid username or password" {
		t.Errorf("msg = %q, want %q", msg, "Invalid username or password")
	}
}

func TestLogout(t *testing.T) {
	ts := newTestServer(t, "/logout.xml", "logout.xml")
	defer ts.Close()
	c := New(ts.URL, ts.Client())

	if err := c.Logout(context.Background(), "1BI22CS123"); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	wantForm(t, ts.form, map[string]string{
		"mode":        "193",
		"username":    "1BI22CS123",
		"producttype": "0",
	})
	wantEpochA(t, ts.form)
	if _, ok := ts.form["password"]; ok {
		t.Error("logout request must not include password")
	}
}

func TestEnvOverride(t *testing.T) {
	ts := newTestServer(t, "/login.xml", "login_live.xml")
	defer ts.Close()
	t.Setenv("CAPTIVE_BYPASS_PORTAL", ts.URL)

	c := New("http://wrong.example:99", nil)
	ok, _, err := c.Login(context.Background(), "1BI22CS123", "hunter2")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
}
