// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     *url.URL
		wantErr bool
		wantURL *url.URL
	}{
		{"BadScheme", &url.URL{Scheme: "bad"}, true, nil},
		{"HTTPBaseURL", &url.URL{Scheme: "http", Host: "p80.pool.sks-keyservers.net"},
			false, &url.URL{Scheme: "http", Host: "p80.pool.sks-keyservers.net"}},
		{"HTTPSBaseURL", &url.URL{Scheme: "https", Host: "hkps.pool.sks-keyservers.net"},
			false, &url.URL{Scheme: "https", Host: "hkps.pool.sks-keyservers.net"}},
		{"HKPBaseURL", &url.URL{Scheme: "hkp", Host: "pool.sks-keyservers.net"},
			false, &url.URL{Scheme: "http", Host: "pool.sks-keyservers.net:11371"}},
		{"HKPSBaseURL", &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net"},
			false, &url.URL{Scheme: "https", Host: "hkps.pool.sks-keyservers.net"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := normalizeURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}

			if err == nil {
				if got, want := u, tt.wantURL; !reflect.DeepEqual(got, want) {
					t.Errorf("got url %v, want %v", got, want)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	httpClient := &http.Client{}

	tests := []struct {
		name           string
		cfg            *Config
		wantErr        bool
		wantURL        string
		wantAuthToken  string
		wantUserAgent  string
		wantHTTPClient *http.Client
	}{
		{"UnsupportedBaseURL", &Config{
			BaseURL: "bad:",
		}, true, "", "", "", nil},
		{"BadBaseURL", &Config{
			BaseURL: ":",
		}, true, "", "", "", nil},
		{"TLSRequiredHTTP", &Config{
			BaseURL:   "http://p80.pool.sks-keyservers.net",
			AuthToken: "blah",
		}, true, "", "", "", nil},
		{"TLSRequiredHKP", &Config{
			BaseURL:   "hkp://pool.sks-keyservers.net",
			AuthToken: "blah",
		}, true, "", "", "", nil},
		{"LocalhostAuthTokenHTTP", &Config{
			BaseURL:   "http://localhost",
			AuthToken: "blah",
		}, false, "http://localhost", "blah", "", http.DefaultClient},
		{"LocalhostAuthTokenHTTP8080", &Config{
			BaseURL:   "http://localhost:8080",
			AuthToken: "blah",
		}, false, "http://localhost:8080", "blah", "", http.DefaultClient},
		{"LocalhostAuthTokenHKP", &Config{
			BaseURL:   "hkp://localhost",
			AuthToken: "blah",
		}, false, "http://localhost:11371", "blah", "", http.DefaultClient},
		{"NilConfig", nil, false, defaultBaseURL, "", "", http.DefaultClient},
		{"BaseURL", &Config{
			BaseURL: "hkps://hkps.pool.sks-keyservers.net",
		}, false, "https://hkps.pool.sks-keyservers.net", "", "", http.DefaultClient},
		{"AuthToken", &Config{
			AuthToken: "blah",
		}, false, defaultBaseURL, "blah", "", http.DefaultClient},
		{"UserAgent", &Config{
			UserAgent: "Secret Agent Man",
		}, false, defaultBaseURL, "", "Secret Agent Man", http.DefaultClient},
		{"HTTPClient", &Config{
			HTTPClient: httpClient,
		}, false, defaultBaseURL, "", "", httpClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}

			if err == nil {
				if got, want := c.BaseURL.String(), tt.wantURL; got != want {
					t.Errorf("got host %v, want %v", got, want)
				}

				if got, want := c.AuthToken, tt.wantAuthToken; got != want {
					t.Errorf("got auth token %v, want %v", got, want)
				}

				if got, want := c.UserAgent, tt.wantUserAgent; got != want {
					t.Errorf("got user agent %v, want %v", got, want)
				}

				if got, want := c.HTTPClient, tt.wantHTTPClient; got != want {
					t.Errorf("got HTTP client %v, want %v", got, want)
				}
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	defaultClient, err := NewClient(DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		client         *Client
		method         string
		path           string
		rawQuery       string
		body           string
		wantErr        bool
		wantURL        string
		wantAuthBearer string
		wantUserAgent  string
	}{
		{"BadMethod", defaultClient, "b@d	", "", "", "", true, "", "", ""},
		{"BadScheme", &Client{
			BaseURL: &url.URL{Scheme: "bad", Host: "localhost"},
		}, http.MethodGet, "/path", "", "", true, "", "", ""},
		{"TLSRequiredHTTP", &Client{
			BaseURL:   &url.URL{Scheme: "http", Host: "p80.pool.sks-keyservers.net"},
			AuthToken: "blah",
		}, "", "", "", "", true, "", "", ""},
		{"TLSRequiredHKP", &Client{
			BaseURL:   &url.URL{Scheme: "hkp", Host: "pool.sks-keyservers.net"},
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", true, "", "", ""},
		{"LocalhostAuthTokenHTTP", &Client{
			BaseURL:   &url.URL{Scheme: "http", Host: "localhost"},
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "http://localhost/path", "BEARER blah", ""},
		{"LocalhostAuthTokenHTTP8080", &Client{
			BaseURL:   &url.URL{Scheme: "http", Host: "localhost:8080"},
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "http://localhost:8080/path", "BEARER blah", ""},
		{"LocalhostAuthTokenHKP", &Client{
			BaseURL:   &url.URL{Scheme: "hkp", Host: "localhost"},
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "http://localhost:11371/path", "BEARER blah", ""},
		{"Get", defaultClient, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "", ""},
		{"Post", defaultClient, http.MethodPost, "/path", "", "", false, "https://keys.sylabs.io/path", "", ""},
		{"PostRawQuery", defaultClient, http.MethodPost, "/path", "a=b", "", false, "https://keys.sylabs.io/path?a=b", "", ""},
		{"PostBody", defaultClient, http.MethodPost, "/path", "", "body", false, "https://keys.sylabs.io/path", "", ""},
		{"BaseURLAbsolute", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net"},
		}, http.MethodGet, "/path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"BaseURLRelative", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net"},
		}, http.MethodGet, "path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"BaseURLPathAbsolute", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net", Path: "/a/b"},
		}, http.MethodGet, "/path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"BaseURLPathRelative", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net", Path: "/a/b"},
		}, http.MethodGet, "path", "", "", false, "https://hkps.pool.sks-keyservers.net/a/path", "", ""},
		{"BaseURLPathSlashAbsolute", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net", Path: "/a/b/"},
		}, http.MethodGet, "/path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"BaseURLPathSlashRelative", &Client{
			BaseURL: &url.URL{Scheme: "hkps", Host: "hkps.pool.sks-keyservers.net", Path: "/a/b/"},
		}, http.MethodGet, "path", "", "", false, "https://hkps.pool.sks-keyservers.net/a/b/path", "", ""},
		{"AuthToken", &Client{
			BaseURL:   defaultClient.BaseURL,
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "BEARER blah", ""},
		{"UserAgent", &Client{
			BaseURL:   defaultClient.BaseURL,
			UserAgent: "Secret Agent Man",
		}, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "", "Secret Agent Man"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := tt.client.newRequest(tt.method, tt.path, tt.rawQuery, strings.NewReader(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if got, want := r.Method, tt.method; got != want {
					t.Errorf("got method %v, want %v", got, want)
				}

				if got, want := r.URL.String(), tt.wantURL; got != want {
					t.Errorf("got URL %v, want %v", got, want)
				}

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read body: %v", err)
				}
				if got, want := string(b), tt.body; got != want {
					t.Errorf("got body %v, want %v", got, want)
				}

				authBearer, ok := r.Header["Authorization"]
				if got, want := ok, (tt.wantAuthBearer != ""); got != want {
					t.Fatalf("presence of auth bearer %v, want %v", got, want)
				}
				if ok {
					if got, want := len(authBearer), 1; got != want {
						t.Fatalf("got %v auth bearer(s), want %v", got, want)
					}
					if got, want := authBearer[0], tt.wantAuthBearer; got != want {
						t.Errorf("got auth bearer %v, want %v", got, want)
					}
				}

				userAgent, ok := r.Header["User-Agent"]
				if got, want := ok, (tt.wantUserAgent != ""); got != want {
					t.Fatalf("presence of user agent %v, want %v", got, want)
				}
				if ok {
					if got, want := len(userAgent), 1; got != want {
						t.Fatalf("got %v user agent(s), want %v", got, want)
					}
					if got, want := userAgent[0], tt.wantUserAgent; got != want {
						t.Errorf("got user agent %v, want %v", got, want)
					}
				}
			}
		})
	}
}
