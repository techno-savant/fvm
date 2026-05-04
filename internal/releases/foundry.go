package releases

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const defaultFoundryBaseURL = "https://foundryvtt.com"

var (
	errFoundryAuthRequired = errors.New("foundry authentication is not configured")
	communityUsernameRE    = regexp.MustCompile(`/community/([^/?#]+)`)
)

var supportedFoundryPlatforms = map[string]struct{}{
	"node":             {},
	"linux":            {},
	"mac":              {},
	"windows":          {},
	"windows_portable": {},
}

// RemoteRelease describes a remotely available Foundry build.
type RemoteRelease struct {
	Version     string
	Channel     string
	DownloadURL string
	Platform    string
}

// Download contains the resolved download metadata for an install.
type Download struct {
	Version   string
	SourceURL string
	Platform  string
}

// HTTPClient is the subset of http.Client used by FoundryProvider.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func roundTripperFromClient(client HTTPClient) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return client.Do(req)
	})
}

// FoundryProvider talks to the foundryvtt.com releases API using cookie auth.
type FoundryProvider struct {
	BaseURL        string
	Cookie         string
	Username       string
	Password       string
	ReleaseChannel string
	Platform       string
	Client         HTTPClient
	Now            func() time.Time
}

type foundryReleaseResponse struct {
	Version  string `json:"version"`
	URL      string `json:"url"`
	Lifetime int    `json:"lifetime"`
}

func NewFoundryProvider(baseURL, cookie, channel string, client HTTPClient) *FoundryProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultFoundryBaseURL
	}
	if strings.TrimSpace(channel) == "" {
		channel = "stable"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &FoundryProvider{
		BaseURL:        strings.TrimRight(baseURL, "/"),
		Cookie:         strings.TrimSpace(cookie),
		ReleaseChannel: strings.TrimSpace(channel),
		Platform:       "node",
		Client:         client,
		Now:            func() time.Time { return time.Now().UTC() },
	}
}

func (p *FoundryProvider) ListRemote() ([]string, error) {
	releases, err := p.ListRemoteReleases()
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(releases))
	for _, rel := range releases {
		versions = append(versions, rel.Version)
	}
	return versions, nil
}

func (p *FoundryProvider) ListRemoteReleases() ([]RemoteRelease, error) {
	return nil, fmt.Errorf("list-remote is not implemented against the authenticated Foundry download API yet")
}

func (p *FoundryProvider) ResolveDownload(version string) (Download, error) {
	platform, err := NormalizeFoundryPlatform(p.Platform)
	if err != nil {
		return Download{}, err
	}
	resp, err := p.fetchReleaseMetadata(version, platform)
	if err != nil {
		return Download{}, err
	}
	resolvedVersion := strings.TrimSpace(resp.Version)
	if resolvedVersion == "" {
		resolvedVersion = version
	}
	if strings.TrimSpace(resp.URL) == "" {
		return Download{}, fmt.Errorf("foundry releases API returned empty download URL for %s", version)
	}
	return Download{
		Version:   resolvedVersion,
		SourceURL: resp.URL,
		Platform:  platform,
	}, nil
}

func (p *FoundryProvider) Install(version, destDir string) error {
	download, err := p.ResolveDownload(version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	if err := p.downloadAndExtract(download.SourceURL, destDir); err != nil {
		return err
	}
	return nil
}

func (p *FoundryProvider) channel() string {
	if strings.TrimSpace(p.ReleaseChannel) == "" {
		return "stable"
	}
	return p.ReleaseChannel
}

func NormalizeFoundryPlatform(value string) (string, error) {
	platform := strings.ToLower(strings.TrimSpace(value))
	if platform == "" {
		platform = "node"
	}
	switch platform {
	case "darwin", "macos", "osx":
		platform = "mac"
	case "win", "win32":
		platform = "windows"
	}
	if _, ok := supportedFoundryPlatforms[platform]; !ok {
		return "", fmt.Errorf("unsupported Foundry platform %q (supported: node, linux, mac, windows, windows_portable)", value)
	}
	return platform, nil
}

func (p *FoundryProvider) authCookie() (string, error) {
	if cookie := strings.TrimSpace(p.Cookie); cookie != "" {
		return cookie, nil
	}
	if strings.TrimSpace(p.Username) == "" || strings.TrimSpace(p.Password) == "" {
		return "", errFoundryAuthRequired
	}
	cookie, err := p.loginAndGetCookie()
	if err != nil {
		return "", err
	}
	p.Cookie = cookie
	return cookie, nil
}

func (p *FoundryProvider) fetchReleaseMetadata(version, platform string) (foundryReleaseResponse, error) {
	cookie, err := p.authCookie()
	if err != nil {
		return foundryReleaseResponse{}, err
	}
	build := versionToBuild(version)
	if build == "" {
		return foundryReleaseResponse{}, fmt.Errorf("invalid Foundry version: %s", version)
	}
	endpoint := fmt.Sprintf("%s/releases/download?build=%s&platform=%s&response_type=json", p.BaseURL, build, platform)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return foundryReleaseResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", cookie)

	resp, err := p.Client.Do(req)
	if err != nil {
		return foundryReleaseResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return foundryReleaseResponse{}, fmt.Errorf("foundry releases API: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var payload foundryReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return foundryReleaseResponse{}, fmt.Errorf("decode foundry releases API response: %w", err)
	}
	return payload, nil
}

func (p *FoundryProvider) loginAndGetCookie() (string, error) {
	var transport http.RoundTripper
	if p.Client != nil {
		transport = roundTripperFromClient(p.Client)
	}
	result, err := performFoundryLogin(p.BaseURL, p.Username, p.Password, transport)
	if err != nil {
		return "", err
	}
	return result.CookieHeader, nil
}

func (p *FoundryProvider) AuthCookieForDebug() (string, error) {
	return p.loginAndGetCookie()
}

func extractCSRFToken(body string) (string, error) {
	formHTML := extractLoginForm(body)
	if strings.TrimSpace(formHTML) == "" {
		return "", fmt.Errorf("could not find login form on Foundry login page")
	}
	token := extractHiddenInputValue(formHTML, "csrfmiddlewaretoken")
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("could not find csrfmiddlewaretoken on Foundry login page")
	}
	return token, nil
}

func extractLoginForm(body string) string {
	re := regexp.MustCompile(`(?is)<form[^>]*\bid="login-form"[^>]*>.*?</form>`)
	return strings.TrimSpace(re.FindString(body))
}

func extractHiddenInputValue(body, name string) string {
	tagPattern := fmt.Sprintf(`(?is)<input[^>]*\bname="%s"[^>]*>`, regexp.QuoteMeta(name))
	tagRE := regexp.MustCompile(tagPattern)
	tag := tagRE.FindString(body)
	if strings.TrimSpace(tag) == "" {
		return ""
	}
	valueRE := regexp.MustCompile(`(?i)\bvalue="([^"]*)"`)
	match := valueRE.FindStringSubmatch(tag)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func extractCommunityUsername(body string) (string, error) {
	match := communityUsernameRE.FindStringSubmatch(body)
	if len(match) < 2 || strings.TrimSpace(match[1]) == "" {
		return "", fmt.Errorf("could not resolve Foundry community username after login")
	}
	return strings.ToLower(strings.TrimSpace(match[1])), nil
}

func collapseSetCookies(values []string) string {
	parts := make([]string, 0, len(values))
	for _, raw := range values {
		segment := strings.TrimSpace(strings.SplitN(raw, ";", 2)[0])
		if segment != "" {
			parts = append(parts, segment)
		}
	}
	return strings.Join(parts, "; ")
}

func mergeCookieHeaders(existing, incoming string) string {
	cookies := map[string]string{}
	for _, header := range []string{existing, incoming} {
		for _, part := range strings.Split(header, ";") {
			piece := strings.TrimSpace(part)
			if piece == "" {
				continue
			}
			kv := strings.SplitN(piece, "=", 2)
			if len(kv) != 2 {
				continue
			}
			cookies[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return cookieMapToHeader(cookies)
}

func cookiesToHeader(cookies []*http.Cookie) string {
	values := make(map[string]string, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil || strings.TrimSpace(cookie.Name) == "" {
			continue
		}
		values[strings.TrimSpace(cookie.Name)] = cookie.Value
	}
	return cookieMapToHeader(values)
}

func cookieMapToHeader(cookies map[string]string) string {
	keys := make([]string, 0, len(cookies))
	for key := range cookies {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+cookies[key])
	}
	return strings.Join(parts, "; ")
}

func (p *FoundryProvider) downloadAndExtract(url, destDir string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if cookie := strings.TrimSpace(p.Cookie); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download foundry release: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	archiveURL, err := urlpkg.Parse(url)
	if err != nil {
		return fmt.Errorf("parse release archive url: %w", err)
	}
	archivePath := strings.ToLower(archiveURL.Path)

	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(resp.Body, destDir)
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		return extractTarGz(resp.Body, destDir)
	default:
		return fmt.Errorf("unsupported release archive format: %s", url)
	}
}

func CurrentPlatform() string {
	return "node"
}

func versionToBuild(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

func extractZip(r io.Reader, destDir string) error {
	tmp, err := os.CreateTemp("", "fvm-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err := io.Copy(tmp, r); err != nil {
		return err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		return err
	}
	info, err := tmp.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(tmp, info.Size())
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		if err := extractZipEntry(f, destDir); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(f *zip.File, destDir string) error {
	target, err := secureJoin(destDir, f.Name)
	if err != nil {
		return err
	}
	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	mode := f.Mode()
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := secureJoin(destDir, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func secureJoin(root, name string) (string, error) {
	cleanRoot := filepath.Clean(root)
	cleanName := filepath.Clean(name)
	if cleanName == "." {
		return cleanRoot, nil
	}
	target := filepath.Join(cleanRoot, cleanName)
	rel, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry escapes destination: %s", name)
	}
	return target, nil
}

func SortVersionsDescending(versions []string) []string {
	out := append([]string(nil), versions...)
	sort.Slice(out, func(i, j int) bool {
		return versionKey(out[i]) > versionKey(out[j])
	})
	return out
}

func versionKey(v string) string {
	parts := strings.Split(v, ".")
	for i, p := range parts {
		parts[i] = fmt.Sprintf("%08s", strings.TrimSpace(p))
	}
	return strings.Join(parts, ".")
}
