// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	jsonresp "github.com/sylabs/json-resp"
)

type MockPKSAdd struct {
	t       *testing.T
	code    int
	keyText string
}

func (m *MockPKSAdd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.code != http.StatusOK {
		if err := jsonresp.WriteError(w, "", m.code); err != nil {
			m.t.Fatalf("failed to write error: %v", err)
		}
		return
	}

	if got, want := r.URL.Path, pathPKSAdd; got != want {
		m.t.Errorf("got path %v, want %v", got, want)
	}

	if got, want := r.Header.Get("Content-Type"), "application/x-www-form-urlencoded"; got != want {
		m.t.Errorf("got content type %v, want %v", got, want)
	}

	if err := r.ParseForm(); err != nil {
		m.t.Fatalf("failed to parse form: %v", err)
	}
	if got, want := r.Form.Get("keytext"), m.keyText; got != want {
		m.t.Errorf("got key text %v, want %v", got, want)
	}
}

func TestPKSAdd(t *testing.T) {
	m := &MockPKSAdd{
		t:       t,
		keyText: "Not valid, but it'll do for testing",
	}
	s := httptest.NewServer(m)
	defer s.Close()

	tests := []struct {
		name    string
		baseURL string
		code    int
	}{
		{"Success", s.URL, http.StatusOK},
		{"JSONError", s.URL, http.StatusBadRequest},
		{"BadURL", "http://127.0.0.1:123456", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code

			c, err := NewClient(&Config{
				BaseURL: tt.baseURL,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			err = c.PKSAdd(context.Background(), m.keyText)

			if tt.code == http.StatusOK {
				if err != nil {
					t.Fatalf("failed to do PKS add: %v", err)
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

type MockPKSLookup struct {
	t             *testing.T
	code          int
	search        string
	op            string
	options       string
	fingerprint   bool
	exact         bool
	pageSize      string
	pageToken     string
	nextPageToken string
	response      string
}

func (m *MockPKSLookup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.code != http.StatusOK {
		if err := jsonresp.WriteError(w, "", m.code); err != nil {
			m.t.Fatalf("failed to write error: %v", err)
		}
		return
	}

	if got, want := r.URL.Path, pathPKSLookup; got != want {
		m.t.Errorf("got path %v, want %v", got, want)
	}

	if got, want := r.ContentLength, int64(0); got != want {
		m.t.Errorf("got content length %v, want %v", got, want)
	}

	if err := r.ParseForm(); err != nil {
		m.t.Fatalf("failed to parse form: %v", err)
	}
	if got, want := r.Form.Get("search"), m.search; got != want {
		m.t.Errorf("got search %v, want %v", got, want)
	}
	if got, want := r.Form.Get("op"), m.op; got != want {
		m.t.Errorf("got op %v, want %v", got, want)
	}
	if got, want := r.Form.Get("options"), m.options; got != want {
		m.t.Errorf("got options %v, want %v", got, want)
	}
	fingerprint := r.Form.Get("fingerprint") == "on"
	if got, want := fingerprint, m.fingerprint; got != want {
		m.t.Errorf("got fingerprint %v, want %v", got, want)
	}
	exact := r.Form.Get("exact") == "on"
	if got, want := exact, m.exact; got != want {
		m.t.Errorf("got exact %v, want %v", got, want)
	}
	if got, want := r.Form.Get("x-pagesize"), m.pageSize; got != want {
		m.t.Errorf("got page size %v, want %v", got, want)
	}
	if got, want := r.Form.Get("x-pagetoken"), m.pageToken; got != want {
		m.t.Errorf("got page token %v, want %v", got, want)
	}

	w.Header().Set("X-HKP-Next-Page-Token", m.nextPageToken)
	io.Copy(w, strings.NewReader(m.response))
}

func TestPKSLookup(t *testing.T) {
	m := &MockPKSLookup{
		t:        t,
		response: "Not valid, but it'll do for testing",
	}
	s := httptest.NewServer(m)
	defer s.Close()

	tests := []struct {
		name          string
		baseURL       string
		code          int
		search        string
		op            string
		options       []string
		fingerprint   bool
		exact         bool
		pageToken     string
		pageSize      int
		nextPageToken string
	}{
		{"Get", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "", 0, ""},
		{"GetNPT", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "", 0, "bar"},
		{"GetSize", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "", 42, ""},
		{"GetSizeNPT", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "", 42, "bar"},
		{"GetPT", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "foo", 0, ""},
		{"GetPTNPT", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "foo", 0, "bar"},
		{"GetPTSize", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "foo", 42, ""},
		{"GetPTSizeNPT", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, false, "foo", 42, "bar"},
		{"GetMachineReadable", s.URL, http.StatusOK, "search", OperationGet, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"GetExact", s.URL, http.StatusOK, "search", OperationGet, []string{}, false, true, "", 0, ""},
		{"Index", s.URL, http.StatusOK, "search", OperationIndex, []string{}, false, false, "", 0, ""},
		{"IndexMachineReadable", s.URL, http.StatusOK, "search", OperationIndex, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"IndexFingerprint", s.URL, http.StatusOK, "search", OperationIndex, []string{}, true, false, "", 0, ""},
		{"IndexExact", s.URL, http.StatusOK, "search", OperationIndex, []string{}, false, true, "", 0, ""},
		{"VIndex", s.URL, http.StatusOK, "search", OperationVIndex, []string{}, false, false, "", 0, ""},
		{"VIndexMachineReadable", s.URL, http.StatusOK, "search", OperationVIndex, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"VIndexFingerprint", s.URL, http.StatusOK, "search", OperationVIndex, []string{}, true, false, "", 0, ""},
		{"VIndexExact", s.URL, http.StatusOK, "search", OperationVIndex, []string{}, false, true, "", 0, ""},
		{"JSONError", s.URL, http.StatusBadRequest, "", "", []string{}, false, false, "", 0, ""},
		{"BadURL", "http://127.0.0.1:123456", 0, "", "", []string{}, false, false, "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code
			m.search = tt.search
			m.op = tt.op
			m.options = strings.Join(tt.options, ",")
			m.fingerprint = tt.fingerprint
			m.exact = tt.exact
			m.pageToken = tt.pageToken
			m.pageSize = strconv.Itoa(tt.pageSize)
			m.nextPageToken = tt.nextPageToken

			c, err := NewClient(&Config{
				BaseURL: tt.baseURL,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			pd := PageDetails{
				Token: tt.pageToken,
				Size:  tt.pageSize,
			}
			r, err := c.PKSLookup(context.Background(), &pd, tt.search, tt.op, tt.fingerprint, tt.exact, tt.options)

			if tt.code == http.StatusOK {
				if err != nil {
					t.Fatalf("failed to do PKS lookup: %v", err)
				}
				if got, want := pd.Token, tt.nextPageToken; got != want {
					t.Errorf("got page token %v, want %v", got, want)
				}
				if got, want := r, m.response; got != want {
					t.Errorf("got response %v, want %v", got, want)
				}
			} else {
				if err == nil {
					t.Fatalf("unexpected success")
				}
				if 0 != tt.code {
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

func TestGetKey(t *testing.T) {
	fp := [20]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13,
	}

	m := &MockPKSLookup{
		t:        t,
		search:   fmt.Sprintf("%#x", fp),
		op:       "get",
		exact:    true,
		response: "Not valid, but it'll do for testing",
	}
	s := httptest.NewServer(m)
	defer s.Close()

	tests := []struct {
		name    string
		baseURL string
		code    int
		fp      [20]byte
	}{
		{"Success", s.URL, http.StatusOK, fp},
		{"JSONError", s.URL, http.StatusBadRequest, fp},
		{"BadURL", "http://127.0.0.1:123456", 0, fp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code

			c, err := NewClient(&Config{
				BaseURL: tt.baseURL,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			kt, err := c.GetKey(context.Background(), tt.fp)

			if tt.code == http.StatusOK {
				if err != nil {
					t.Fatalf("failed to get key: %v", err)
				}
				if got, want := kt, m.response; got != want {
					t.Errorf("got keyText %v, want %v", got, want)
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
