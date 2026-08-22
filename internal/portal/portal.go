package portal

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://rr.pes.edu:8090"

type Client struct {
	baseURL string
	client  *http.Client
}

type portalResponse struct {
	Status  string `xml:"status"`
	Message string `xml:"message"`
}

func New(baseURL string, hc *http.Client) *Client {
	if env := os.Getenv("CAPTIVE_BYPASS_PORTAL"); env != "" {
		baseURL = env
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if hc == nil {
		hc = &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), client: hc}
}

func (c *Client) Login(ctx context.Context, username, password string) (bool, string, error) {
	form := url.Values{}
	form.Set("mode", "191")
	form.Set("username", username)
	form.Set("password", password)
	form.Set("a", strconv.FormatInt(time.Now().UnixMilli(), 10))
	form.Set("producttype", "0")

	respBody, err := c.post(ctx, "/login.xml", form)
	if err != nil {
		return false, "", err
	}

	var pr portalResponse
	if err := xml.Unmarshal(respBody, &pr); err != nil {
		return false, "", err
	}
	if pr.Status == "LIVE" {
		return true, "authenticated", nil
	}
	return false, pr.Message, nil
}

func (c *Client) Logout(ctx context.Context, username string) error {
	form := url.Values{}
	form.Set("mode", "193")
	form.Set("username", username)
	form.Set("a", strconv.FormatInt(time.Now().UnixMilli(), 10))
	form.Set("producttype", "0")

	_, err := c.post(ctx, "/logout.xml", form)
	return err
}

func (c *Client) Livecheck(ctx context.Context) (bool, error) {
	respBody, err := c.post(ctx, "/livecheck.xml", url.Values{})
	if err != nil {
		return false, err
	}
	var pr portalResponse
	if err := xml.Unmarshal(respBody, &pr); err != nil {
		return false, err
	}
	return pr.Status == "LIVE", nil
}

func (c *Client) post(ctx context.Context, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "captive-bypass/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
