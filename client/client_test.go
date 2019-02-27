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
		wantHost       string
		wantAuthToken  string
		wantUserAgent  string
		wantHTTPClient *http.Client
	}{
		{"NilConfig", nil, false, "keys.sylabs.io", "", "", http.DefaultClient},
		{"BaseURL", &Config{
			BaseURL: "https://keys.staging.sycloud.io",
		}, false, "keys.staging.sycloud.io", "", "", http.DefaultClient},
		{"BadBaseURL", &Config{
			BaseURL: ":",
		}, true, "", "", "", nil},
		{"AuthToken", &Config{
			AuthToken: "blah",
		}, false, "keys.sylabs.io", "blah", "", http.DefaultClient},
		{"UserAgent", &Config{
			UserAgent: "Secret Agent Man",
		}, false, "keys.sylabs.io", "", "Secret Agent Man", http.DefaultClient},
		{"HTTPClient", &Config{
			HTTPClient: httpClient,
		}, false, "keys.sylabs.io", "", "", httpClient},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}

			if err == nil {
				if got, want := c.baseURL.Host, tt.wantHost; got != want {
					t.Errorf("got host %v, want %v", got, want)
				}

				if got, want := c.authToken, tt.wantAuthToken; got != want {
					t.Errorf("got auth token %v, want %v", got, want)
				}

				if got, want := c.userAgent, tt.wantUserAgent; got != want {
					t.Errorf("got user agent %v, want %v", got, want)
				}

				if got, want := c.httpClient, tt.wantHTTPClient; got != want {
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
		wantAuthBearer string
		wantUserAgent  string
	}{
		{"BadMethod", nil, "b@d	", "", "", "", true, "", ""},
		{"NilConfigGet", nil, http.MethodGet, "/path", "", "", false, "", ""},
		{"NilConfigPost", nil, http.MethodPost, "/path", "", "", false, "", ""},
		{"NilConfigPostRawQuery", nil, http.MethodPost, "/path", "a=b", "", false, "", ""},
		{"NilConfigPostBody", nil, http.MethodPost, "/path", "", "body", false, "", ""},
		{"BaseURL", &Config{
			BaseURL: "https://keys.staging.sycloud.io",
		}, http.MethodGet, "/path", "", "", false, "", ""},
		{"AuthToken", &Config{
			AuthToken: "blah",
		}, http.MethodGet, "/path", "", "", false, "BEARER blah", ""},
		{"UserAgent", &Config{
			UserAgent: "Secret Agent Man",
		}, http.MethodGet, "/path", "", "", false, "", "Secret Agent Man"},
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

				if got, want := r.URL.Path, tt.path; got != want {
					t.Errorf("got path %v, want %v", got, want)
				}

				if got, want := r.URL.RawQuery, tt.rawQuery; got != want {
					t.Errorf("got query %v, want %v", got, want)
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
