package releases_test

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/foundry/fvm/internal/install"
	"github.com/foundry/fvm/internal/releases"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListInstalled_empty_when_no_versions_dir(t *testing.T) {
	dir := t.TempDir()
	vs, err := releases.ListInstalled(filepath.Join(dir, "versions"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(vs))
	}
}

func TestListInstalled_skips_incomplete(t *testing.T) {
	dir := t.TempDir()
	vDir := filepath.Join(dir, "13.346")

	if err := install.WriteManifest(vDir, install.Manifest{Version: "13.346"}); err != nil {
		t.Fatal(err)
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("expected empty list (incomplete install), got %d", len(vs))
	}
}

func TestListInstalled_includes_complete(t *testing.T) {
	dir := t.TempDir()
	vDir := filepath.Join(dir, "13.346")
	l := install.NewLayout(vDir)

	if err := install.WriteManifest(vDir, install.Manifest{Version: "13.346"}); err != nil {
		t.Fatal(err)
	}
	if err := l.MarkComplete(); err != nil {
		t.Fatal(err)
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 1 {
		t.Fatalf("expected 1 installed version, got %d", len(vs))
	}
	if vs[0].Version != "13.346" {
		t.Fatalf("expected 13.346, got %q", vs[0].Version)
	}
}

func TestListInstalled_multiple_versions(t *testing.T) {
	dir := t.TempDir()
	for _, ver := range []string{"12.345", "13.346"} {
		vDir := filepath.Join(dir, ver)
		l := install.NewLayout(vDir)
		if err := install.WriteManifest(vDir, install.Manifest{Version: ver}); err != nil {
			t.Fatal(err)
		}
		if err := l.MarkComplete(); err != nil {
			t.Fatal(err)
		}
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(vs))
	}
}

func TestStubProvider_list_remote(t *testing.T) {
	p := &releases.StubProvider{Versions: []string{"13.346", "12.345"}}
	vs, err := p.ListRemote()
	if err != nil {
		t.Fatalf("ListRemote: %v", err)
	}
	if len(vs) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(vs))
	}
}

func TestStubProvider_install_returns_error(t *testing.T) {
	p := &releases.StubProvider{}
	if err := p.Install("13.346", t.TempDir()); err == nil {
		t.Fatal("expected error from stub installer")
	}
}

func TestNewInstallRecord_creates_tmp_dir(t *testing.T) {
	dir := t.TempDir()
	versionsRoot := filepath.Join(dir, "versions")
	tmpBase := filepath.Join(dir, "tmp")

	rec, err := releases.NewInstallRecord("13.346", versionsRoot, tmpBase)
	if err != nil {
		t.Fatalf("NewInstallRecord: %v", err)
	}
	if rec.TmpDir == "" {
		t.Fatal("expected non-empty TmpDir")
	}
}

func TestInstallRecord_finalize(t *testing.T) {
	dir := t.TempDir()
	versionsRoot := filepath.Join(dir, "versions")
	tmpBase := filepath.Join(dir, "tmp")

	rec, err := releases.NewInstallRecord("13.346", versionsRoot, tmpBase)
	if err != nil {
		t.Fatal(err)
	}

	m := install.Manifest{
		Version:     "13.346",
		Platform:    "linux",
		Arch:        "amd64",
		InstalledAt: time.Now().UTC(),
	}
	if err := rec.Finalize(m); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if !rec.Layout.IsComplete() {
		t.Fatal("expected install to be complete after Finalize")
	}

	got, err := install.ReadManifest(rec.Layout.Root)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}
	if got.Version != "13.346" {
		t.Fatalf("expected 13.346, got %q", got.Version)
	}
}

func TestInstallRecord_cleanup(t *testing.T) {
	dir := t.TempDir()
	rec, err := releases.NewInstallRecord("13.346", filepath.Join(dir, "versions"), filepath.Join(dir, "tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rec.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
}

func TestIsVersionString(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		{"13.346", true},
		{"12.345", true},
		{"1.0.0", true},
		{"11", true},
		{"", false},
		{"abc", false},
		{"13.346-beta", false},
		{"  ", false},
	}
	for _, c := range cases {
		if got := releases.IsVersionString(c.input); got != c.valid {
			t.Errorf("IsVersionString(%q) = %v, want %v", c.input, got, c.valid)
		}
	}
}

func TestCurrentPlatformDefault(t *testing.T) {
	if got := releases.CurrentPlatform(); got != "node" {
		t.Fatalf("expected default platform node, got %q", got)
	}
}

func TestNormalizeFoundryPlatform(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", "node"},
		{"node", "node"},
		{"darwin", "mac"},
		{"macos", "mac"},
		{"linux", "linux"},
		{"win", "windows"},
		{"windows_portable", "windows_portable"},
	}
	for _, tc := range cases {
		got, err := releases.NormalizeFoundryPlatform(tc.input)
		if err != nil {
			t.Fatalf("NormalizeFoundryPlatform(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeFoundryPlatform(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeFoundryPlatform_rejectsUnknown(t *testing.T) {
	if _, err := releases.NormalizeFoundryPlatform("plan9"); err == nil {
		t.Fatal("expected unsupported platform error")
	}
}

func TestFoundryProviderResolveDownload_usesNodeByDefault(t *testing.T) {
	provider := releases.NewFoundryProvider("https://foundryvtt.com", "sessionid=abc; csrftoken=def", "stable", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/releases/download" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		q := req.URL.Query()
		if q.Get("build") != "360" {
			t.Fatalf("expected build=360, got %q", q.Get("build"))
		}
		if q.Get("platform") != "node" {
			t.Fatalf("expected platform=node, got %q", q.Get("platform"))
		}
		if q.Get("response_type") != "json" {
			t.Fatalf("expected response_type=json, got %q", q.Get("response_type"))
		}
		if !strings.Contains(req.Header.Get("Cookie"), "sessionid=abc") {
			t.Fatalf("expected Cookie header, got %q", req.Header.Get("Cookie"))
		}
		body := `{"version":"14.360","url":"https://r2.foundryvtt.com/releases/14.360/FoundryVTT-14.360.zip?verify=abc","lifetime":300}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	download, err := provider.ResolveDownload("14.360")
	if err != nil {
		t.Fatalf("ResolveDownload: %v", err)
	}
	if download.Version != "14.360" {
		t.Fatalf("expected version 14.360, got %q", download.Version)
	}
	if !strings.Contains(download.SourceURL, "FoundryVTT-14.360.zip") {
		t.Fatalf("unexpected source URL: %q", download.SourceURL)
	}
	if download.Platform != "node" {
		t.Fatalf("expected platform node, got %q", download.Platform)
	}
}

func TestFoundryProviderResolveDownload_normalizesConfiguredPlatform(t *testing.T) {
	provider := releases.NewFoundryProvider("https://foundryvtt.com", "sessionid=abc; csrftoken=def", "stable", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.Query().Get("platform"); got != "mac" {
			t.Fatalf("expected normalized platform mac, got %q", got)
		}
		body := `{"version":"14.360","url":"https://r2.foundryvtt.com/releases/14.360/FoundryVTT-14.360.dmg?verify=abc","lifetime":300}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))
	provider.Platform = "darwin"

	download, err := provider.ResolveDownload("14.360")
	if err != nil {
		t.Fatalf("ResolveDownload: %v", err)
	}
	if download.Platform != "mac" {
		t.Fatalf("expected normalized platform mac, got %q", download.Platform)
	}
}

func TestFoundryProviderResolveDownload_requiresAuth(t *testing.T) {
	provider := releases.NewFoundryProvider("https://foundryvtt.com", "", "stable", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, nil
	}))
	if _, err := provider.ResolveDownload("14.360"); err == nil {
		t.Fatal("expected auth error")
	}
}

func TestFoundryProviderInstall_rejectsUnsupportedArchiveFormat(t *testing.T) {
	provider := releases.NewFoundryProvider("https://foundryvtt.com", "sessionid=abc", "stable", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Host, "foundryvtt.com") {
			body := `{"version":"14.360","url":"https://r2.foundryvtt.com/releases/14.360/FoundryVTT-14.360.dmg?verify=abc","lifetime":300}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	}))
	provider.Platform = "mac"
	if err := provider.Install("14.360", t.TempDir()); err == nil {
		t.Fatal("expected unsupported archive format error for dmg")
	}
}

func TestFoundryProviderResolveDownload_logsInWithCredentials(t *testing.T) {
	var requests []string
	provider := releases.NewFoundryProvider("https://foundryvtt.com", "", "stable", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.Method+" "+req.URL.String())
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/auth/login/":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`<html><form id="login-form" method="post" action="/auth/login/"><input type="hidden" name="csrfmiddlewaretoken" value="csrf123"><input type="hidden" value="/auth/login/" name="next"></form></html>`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		case req.Method == http.MethodPost && req.URL.Path == "/auth/login/":
			if got := req.FormValue("next"); got != "/auth/login/" {
				t.Fatalf("expected next=/auth/login/, got %q", got)
			}
			if _, ok := req.PostForm["login"]; !ok {
				t.Fatal("expected login field in form submission")
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`<div id="login-welcome"><a href="/community/SavantUser">profile</a></div>`)),
				Header:     make(http.Header),
				Request:    req,
			}
			resp.Header.Add("Set-Cookie", "csrftoken=csrf123; Path=/")
			resp.Header.Add("Set-Cookie", "sessionid=session456; Path=/")
			return resp, nil
		case req.Method == http.MethodGet && req.URL.Path == "/releases/download":
			cookie := req.Header.Get("Cookie")
			if !strings.Contains(cookie, "sessionid=session456") {
				t.Fatalf("expected session cookie after login, got %q", cookie)
			}
			body := `{"version":"14.360","url":"https://r2.foundryvtt.com/releases/14.360/FoundryVTT-14.360.zip?verify=abc","lifetime":300}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	}))
	provider.Username = "savant@example.com"
	provider.Password = "secret"

	download, err := provider.ResolveDownload("14.360")
	if err != nil {
		t.Fatalf("ResolveDownload: %v", err)
	}
	if download.Platform != "node" {
		t.Fatalf("expected platform node, got %q", download.Platform)
	}
	if provider.Cookie == "" {
		t.Fatal("expected provider cookie to be populated after login")
	}
	if len(requests) != 3 {
		t.Fatalf("expected 3 requests (GET login, POST login, GET release), got %d", len(requests))
	}
}
