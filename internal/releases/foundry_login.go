package releases

// shared login helper used by both provider code and standalone repros

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type foundryLoginResult struct {
	CookieHeader string
	FinalURL     string
	ResponseBody string
}

func performFoundryLogin(baseURL, username, password string, transport http.RoundTripper) (foundryLoginResult, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return foundryLoginResult{}, errFoundryAuthRequired
	}

	baseURL = strings.TrimRight(baseURL, "/")
	loginURL := baseURL + "/auth/login/"
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return foundryLoginResult{}, fmt.Errorf("parse Foundry base URL: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return foundryLoginResult{}, fmt.Errorf("create cookie jar: %w", err)
	}
	client := &http.Client{Jar: jar}
	if transport != nil {
		client.Transport = transport
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		return nil
	}

	getReq, err := http.NewRequest(http.MethodGet, loginURL, nil)
	if err != nil {
		return foundryLoginResult{}, err
	}
	getReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	getReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	getReq.Header.Set("Cache-Control", "max-age=0")
	getReq.Header.Set("Upgrade-Insecure-Requests", "1")
	getReq.Header.Set("Sec-Fetch-Dest", "document")
	getReq.Header.Set("Sec-Fetch-Mode", "navigate")
	getReq.Header.Set("Sec-Fetch-Site", "none")
	getReq.Header.Set("Sec-Fetch-User", "?1")
	getReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")

	getResp, err := client.Do(getReq)
	if err != nil {
		return foundryLoginResult{}, fmt.Errorf("request Foundry login page: %w", err)
	}
	body, readErr := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	if readErr != nil {
		return foundryLoginResult{}, fmt.Errorf("read Foundry login page: %w", readErr)
	}
	if getResp.StatusCode != http.StatusOK {
		return foundryLoginResult{}, fmt.Errorf("request Foundry login page: %s", getResp.Status)
	}

	formHTML := extractLoginForm(string(body))
	if strings.TrimSpace(formHTML) == "" {
		return foundryLoginResult{}, fmt.Errorf("could not find login form on Foundry login page")
	}
	csrfToken := extractHiddenInputValue(formHTML, "csrfmiddlewaretoken")
	if strings.TrimSpace(csrfToken) == "" {
		return foundryLoginResult{}, fmt.Errorf("could not find csrfmiddlewaretoken on Foundry login page")
	}
	nextVal := extractHiddenInputValue(formHTML, "next")
	if strings.TrimSpace(nextVal) == "" {
		nextVal = "/"
	}

	form := url.Values{}
	form.Set("csrfmiddlewaretoken", csrfToken)
	form.Set("next", nextVal)
	form.Set("username", username)
	form.Set("password", password)
	form.Set("login", "")

	formBody := form.Encode()
	postReq, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBufferString(formBody))
	if err != nil {
		return foundryLoginResult{}, err
	}
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("Origin", baseURL)
	postReq.Header.Set("Referer", baseURL+"/")
	postReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	postReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	postReq.Header.Set("Cache-Control", "max-age=0")
	postReq.Header.Set("DNT", "1")
	postReq.Header.Set("Upgrade-Insecure-Requests", "1")
	postReq.Header.Set("Sec-Fetch-Dest", "document")
	postReq.Header.Set("Sec-Fetch-Mode", "navigate")
	postReq.Header.Set("Sec-Fetch-Site", "same-origin")
	postReq.Header.Set("Sec-Fetch-User", "?1")
	postReq.Header.Set("sec-ch-ua", `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`)
	postReq.Header.Set("sec-ch-ua-mobile", "?0")
	postReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
	postReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")

	postResp, err := client.Do(postReq)
	if err != nil {
		return foundryLoginResult{}, fmt.Errorf("submit Foundry login form: %w", err)
	}
	postBody, readErr := io.ReadAll(postResp.Body)
	postResp.Body.Close()
	if readErr != nil {
		return foundryLoginResult{}, fmt.Errorf("read Foundry login response: %w", readErr)
	}
	if postResp.StatusCode != http.StatusOK && postResp.StatusCode != http.StatusFound {
		bodyPreview := strings.TrimSpace(string(postBody))
		if len(bodyPreview) > 300 {
			bodyPreview = bodyPreview[:300]
		}
		return foundryLoginResult{}, fmt.Errorf("submit Foundry login form: %s: %s", postResp.Status, bodyPreview)
	}

	cookieHeader := cookiesToHeader(jar.Cookies(parsedBaseURL))
	responseBody := string(postBody)
	finalURL := postResp.Request.URL.String()
	if postResp.StatusCode == http.StatusFound {
		location := postResp.Header.Get("Location")
		if location != "" {
			redirectURL := location
			if parsedLocation, parseErr := postResp.Request.URL.Parse(location); parseErr == nil {
				redirectURL = parsedLocation.String()
				finalURL = redirectURL
				redirectReq, reqErr := http.NewRequest(http.MethodGet, redirectURL, nil)
				if reqErr == nil {
					redirectReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
					redirectReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
					redirectReq.Header.Set("Cache-Control", "max-age=0")
					redirectReq.Header.Set("DNT", "1")
					redirectReq.Header.Set("Referer", baseURL+"/")
					redirectReq.Header.Set("Upgrade-Insecure-Requests", "1")
					redirectReq.Header.Set("Sec-Fetch-Dest", "document")
					redirectReq.Header.Set("Sec-Fetch-Mode", "navigate")
					redirectReq.Header.Set("Sec-Fetch-Site", "same-origin")
					redirectReq.Header.Set("Sec-Fetch-User", "?1")
					redirectReq.Header.Set("sec-ch-ua", `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`)
					redirectReq.Header.Set("sec-ch-ua-mobile", "?0")
					redirectReq.Header.Set("sec-ch-ua-platform", `"macOS"`)
					redirectReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
					if redirectResp, redirectErr := client.Do(redirectReq); redirectErr == nil {
						redirectBody, _ := io.ReadAll(redirectResp.Body)
						redirectResp.Body.Close()
						responseBody = string(redirectBody)
						finalURL = redirectResp.Request.URL.String()
						cookieHeader = cookiesToHeader(jar.Cookies(parsedBaseURL))
					}
				}
			}
		}
	}
	result := foundryLoginResult{
		CookieHeader: cookieHeader,
		FinalURL:     finalURL,
		ResponseBody: responseBody,
	}
	if strings.Contains(cookieHeader, "sessionid=") {
		return result, nil
	}

	postText := strings.ToLower(result.ResponseBody)
	if strings.Contains(postText, "invalid login credentials") || strings.Contains(postText, "invalid username or password") {
		return foundryLoginResult{}, fmt.Errorf("unable to log in to Foundry with supplied credentials")
	}
	if strings.Contains(postText, `name="username"`) || strings.Contains(postText, `name="password"`) {
		return foundryLoginResult{}, fmt.Errorf("Foundry username/password login was rejected and the login form was re-rendered without a session cookie. Use FOUNDRY_COOKIE with a valid browser session instead")
	}

	return foundryLoginResult{}, fmt.Errorf("Foundry username/password login did not produce a session cookie. Use FOUNDRY_COOKIE with a valid browser session instead")
}
