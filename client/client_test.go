// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

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
		{"NilConfig", nil, false, defaultBaseURL, "", "", http.DefaultClient},
		{"HTTPBaseURL", &Config{
			BaseURL: "http://p80.pool.sks-keyservers.net",
		}, false, "http://p80.pool.sks-keyservers.net", "", "", http.DefaultClient},
		{"HTTPSBaseURL", &Config{
			BaseURL: "https://hkps.pool.sks-keyservers.net",
		}, false, "https://hkps.pool.sks-keyservers.net", "", "", http.DefaultClient},
		{"HKPBaseURL", &Config{
			BaseURL: "hkp://pool.sks-keyservers.net",
		}, false, "http://pool.sks-keyservers.net:11371", "", "", http.DefaultClient},
		{"HKPSBaseURL", &Config{
			BaseURL: "hkps://hkps.pool.sks-keyservers.net",
		}, false, "https://hkps.pool.sks-keyservers.net", "", "", http.DefaultClient},
		{"UnsupportedBaseURL", &Config{
			BaseURL: "bad:",
		}, true, "", "", "", nil},
		{"BadBaseURL", &Config{
			BaseURL: ":",
		}, true, "", "", "", nil},
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
	tests := []struct {
		name           string
		cfg            *Config
		method         string
		path           string
		rawQuery       string
		body           string
		wantErr        bool
		wantURL        string
		wantAuthBearer string
		wantUserAgent  string
	}{
		{"BadMethod", nil, "b@d	", "", "", "", true, "", "", ""},
		{"NilConfigGet", nil, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "", ""},
		{"NilConfigPost", nil, http.MethodPost, "/path", "", "", false, "https://keys.sylabs.io/path", "", ""},
		{"NilConfigPostRawQuery", nil, http.MethodPost, "/path", "a=b", "", false, "https://keys.sylabs.io/path?a=b", "", ""},
		{"NilConfigPostBody", nil, http.MethodPost, "/path", "", "body", false, "https://keys.sylabs.io/path", "", ""},
		{"HTTPBaseURL", &Config{
			BaseURL: "http://p80.pool.sks-keyservers.net",
		}, http.MethodGet, "/path", "", "", false, "http://p80.pool.sks-keyservers.net/path", "", ""},
		{"HTTPSBaseURL", &Config{
			BaseURL: "https://hkps.pool.sks-keyservers.net",
		}, http.MethodGet, "/path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"HKPBaseURL", &Config{
			BaseURL: "hkp://pool.sks-keyservers.net",
		}, http.MethodGet, "/path", "", "", false, "http://pool.sks-keyservers.net:11371/path", "", ""},
		{"HKPSBaseURL", &Config{
			BaseURL: "hkps://hkps.pool.sks-keyservers.net",
		}, http.MethodGet, "/path", "", "", false, "https://hkps.pool.sks-keyservers.net/path", "", ""},
		{"AuthToken", &Config{
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "BEARER blah", ""},
		{"UserAgent", &Config{
			UserAgent: "Secret Agent Man",
		}, http.MethodGet, "/path", "", "", false, "https://keys.sylabs.io/path", "", "Secret Agent Man"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.cfg)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			r, err := c.newRequest(tt.method, tt.path, tt.rawQuery, strings.NewReader(tt.body))
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
