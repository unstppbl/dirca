package dirber

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

// NewHTTPClient returns a new HTTPClient
func (d *Dirber) newHTTPClient() (client *http.Client) {
	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !d.opts.FollowRedirect {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		redirectFunc = nil
	}
	client = &http.Client{
		Timeout:       d.opts.Timeout,
		CheckRedirect: redirectFunc,
		Transport:     d.transport,
	}
	return client
}

// NewHTTPTransport returns transport based on options
func NewHTTPTransport(opt *Options) (*http.Transport, error) {
	var proxyURLFunc func(*http.Request) (*url.URL, error)
	proxyURLFunc = http.ProxyFromEnvironment

	if opt == nil {
		return nil, fmt.Errorf("options is nil")
	}

	if opt.Proxy != "" {
		proxyURL, err := url.Parse(opt.Proxy)
		if err != nil {
			return nil, fmt.Errorf("proxy URL is invalid (%v)", err)
		}
		proxyURLFunc = http.ProxyURL(proxyURL)
	}

	tr := &http.Transport{
		Proxy: proxyURLFunc,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opt.InsecureSSL,
		},
	}
	return tr, nil
}

// MakeRequest makes a request to the specified url
func (d *Dirber) makeRequest(fullURL string) (*int, *int64, error) {
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, nil, err
	}
	client := d.newHTTPClient()
	ua := uaGens[rand.Intn(len(uaGens))]()
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, nil, fmt.Errorf("Invalid certificate: %v", ue.Err)
			}
		}
		return nil, nil, err
	}

	defer resp.Body.Close()

	var length *int64

	if d.opts.IncludeLength {
		length = new(int64)
		if resp.ContentLength <= 0 {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 == nil {
				*length = int64(utf8.RuneCountInString(string(body)))
			}
		} else {
			*length = resp.ContentLength
		}
	} else {
		// DO NOT REMOVE!
		// absolutely needed so golang will reuse connections!
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return nil, nil, err
		}
	}
	return &resp.StatusCode, length, nil
}
