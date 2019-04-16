// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jsonresp "github.com/sylabs/json-resp"
)

type MockVersion struct {
	t       *testing.T
	code    int
	version string
}

func (m *MockVersion) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.code != http.StatusOK {
		if err := jsonresp.WriteError(w, "", m.code); err != nil {
			m.t.Fatalf("failed to write error: %v", err)
		}
		return
	}

	if got, want := r.URL.Path, pathVersion; got != want {
		m.t.Errorf("got path %v, want %v", got, want)
	}

	if got, want := r.ContentLength, int64(0); got != want {
		m.t.Errorf("got content length %v, want %v", got, want)
	}

	vi := VersionInfo{
		Version: m.version,
	}
	if err := jsonresp.WriteResponse(w, vi, m.code); err != nil {
		m.t.Fatalf("failed to write response: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	m := &MockVersion{
		t: t,
	}
	s := httptest.NewServer(m)
	defer s.Close()

	tests := []struct {
		name    string
		baseURL string
		code    int
		version string
	}{
		{"Success", s.URL, http.StatusOK, "1.2.3"},
		{"JSONError", s.URL, http.StatusBadRequest, ""},
		{"BadURL", "http://127.0.0.1:123456", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code
			m.version = tt.version

			c, err := NewClient(&Config{
				BaseURL: tt.baseURL,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			vi, err := c.GetVersion(context.Background())

			if tt.code == http.StatusOK {
				if err != nil {
					t.Fatalf("failed to get version: %v", err)
				}
				if got, want := vi.Version, tt.version; got != want {
					t.Errorf("got version %v, want %v", got, want)
				}
			} else {
				if err == nil {
					t.Fatalf("unexpected success")
				}
				if tt.code != 0 {
					if err, ok := err.(*jsonresp.Error); !ok {
						t.Fatalf("failed to cast to jsonresp.Error")
					} else if got, want := err.Code, tt.code; got != want {
						t.Errorf("got code %v, want %v", got, want)
					}
				}
			}
		})
	}
}
