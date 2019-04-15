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
	message string
	keyText string
}

func (m *MockPKSAdd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.code != http.StatusOK {
		if m.message != "" {
			if err := jsonresp.WriteError(w, m.message, m.code); err != nil {
				m.t.Fatalf("failed to write error: %v", err)
			}
		} else {
			w.WriteHeader(m.code)
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
		t: t,
	}
	s := httptest.NewServer(m)
	defer s.Close()

	tests := []struct {
		name    string
		baseURL string
		keyText string
		code    int
		message string
	}{
		{"Success", s.URL, "key", http.StatusOK, ""},
		{"Error", s.URL, "key", http.StatusBadRequest, ""},
		{"ErrorMessage", s.URL, "key", http.StatusBadRequest, "blah"},
		{"BadURL", "http://127.0.0.1:123456", "key", 0, ""},
		{"InvalidKeyText", s.URL, "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code
			m.message = tt.message
			m.keyText = tt.keyText

			c, err := NewClient(&Config{
				BaseURL: tt.baseURL,
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			err = c.PKSAdd(context.Background(), tt.keyText)

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
					} else if got, want := err.Message, tt.message; got != want {
						t.Errorf("got message %v, want %v", got, want)
					}
				}
			}
		})
	}
}

type MockPKSLookup struct {
	t             *testing.T
	code          int
	message       string
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
		if m.message != "" {
			if err := jsonresp.WriteError(w, m.message, m.code); err != nil {
				m.t.Fatalf("failed to write error: %v", err)
			}
		} else {
			w.WriteHeader(m.code)
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

	// options is optional.
	options, ok := r.Form["options"]
	if got, want := ok, m.options != ""; got != want {
		m.t.Errorf("options presence %v, want %v", got, want)
	} else if ok {
		if len(options) != 1 {
			m.t.Errorf("got multiple options values")
		} else if got, want := options[0], m.options; got != want {
			m.t.Errorf("got options %v, want %v", got, want)
		}
	}

	// fingerprint is optional.
	fp, ok := r.Form["fingerprint"]
	if got, want := ok, m.fingerprint; got != want {
		m.t.Errorf("fingerprint presence %v, want %v", got, want)
	} else if ok {
		if len(fp) != 1 {
			m.t.Errorf("got multiple fingerprint values")
		} else if got, want := fp[0], "on"; got != want {
			m.t.Errorf("got fingerprint %v, want %v", got, want)
		}
	}

	// exact is optional.
	exact, ok := r.Form["exact"]
	if got, want := ok, m.exact; got != want {
		m.t.Errorf("exact presence %v, want %v", got, want)
	} else if ok {
		if len(exact) != 1 {
			m.t.Errorf("got multiple exact values")
		} else if got, want := exact[0], "on"; got != want {
			m.t.Errorf("got exact %v, want %v", got, want)
		}
	}

	// x-pagesize is optional.
	pageSize, ok := r.Form["x-pagesize"]
	if got, want := ok, m.pageSize != ""; got != want {
		m.t.Errorf("page size presence %v, want %v", got, want)
	} else if ok {
		if len(pageSize) != 1 {
			m.t.Error("got multiple page size values")
		} else if got, want := pageSize[0], m.pageSize; got != want {
			m.t.Errorf("got page size %v, want %v", got, want)
		}
	}

	// x-pagetoken is optional.
	pageToken, ok := r.Form["x-pagetoken"]
	if got, want := ok, m.pageToken != ""; got != want {
		m.t.Errorf("page token presence %v, want %v", got, want)
	} else if ok {
		if len(pageToken) != 1 {
			m.t.Error("got multiple page token values")
		} else if got, want := pageToken[0], m.pageToken; got != want {
			m.t.Errorf("got page token %v, want %v", got, want)
		}
	}

	w.Header().Set("X-HKP-Next-Page-Token", m.nextPageToken)
	if _, err := io.Copy(w, strings.NewReader(m.response)); err != nil {
		m.t.Fatalf("failed to copy: %v", err)
	}
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
		message       string
		search        string
		op            string
		options       []string
		fingerprint   bool
		exact         bool
		pageToken     string
		pageSize      int
		nextPageToken string
	}{
		{"Get", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "", 0, ""},
		{"GetNPT", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "", 0, "bar"},
		{"GetSize", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "", 42, ""},
		{"GetSizeNPT", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "", 42, "bar"},
		{"GetPT", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "foo", 0, ""},
		{"GetPTNPT", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "foo", 0, "bar"},
		{"GetPTSize", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "foo", 42, ""},
		{"GetPTSizeNPT", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, false, "foo", 42, "bar"},
		{"GetMachineReadable", s.URL, http.StatusOK, "", "search", OperationGet, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"GetMachineReadableBlah", s.URL, http.StatusOK, "", "search", OperationGet, []string{OptionMachineReadable, "blah"}, false, false, "", 0, ""},
		{"GetExact", s.URL, http.StatusOK, "", "search", OperationGet, []string{}, false, true, "", 0, ""},
		{"Index", s.URL, http.StatusOK, "", "search", OperationIndex, []string{}, false, false, "", 0, ""},
		{"IndexMachineReadable", s.URL, http.StatusOK, "", "search", OperationIndex, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"IndexMachineReadableBlah", s.URL, http.StatusOK, "", "search", OperationIndex, []string{OptionMachineReadable, "blah"}, false, false, "", 0, ""},
		{"IndexFingerprint", s.URL, http.StatusOK, "", "search", OperationIndex, []string{}, true, false, "", 0, ""},
		{"IndexExact", s.URL, http.StatusOK, "", "search", OperationIndex, []string{}, false, true, "", 0, ""},
		{"VIndex", s.URL, http.StatusOK, "", "search", OperationVIndex, []string{}, false, false, "", 0, ""},
		{"VIndexMachineReadable", s.URL, http.StatusOK, "", "search", OperationVIndex, []string{OptionMachineReadable}, false, false, "", 0, ""},
		{"VIndexMachineReadableBlah", s.URL, http.StatusOK, "", "search", OperationVIndex, []string{OptionMachineReadable, "blah"}, false, false, "", 0, ""},
		{"VIndexFingerprint", s.URL, http.StatusOK, "", "search", OperationVIndex, []string{}, true, false, "", 0, ""},
		{"VIndexExact", s.URL, http.StatusOK, "", "search", OperationVIndex, []string{}, false, true, "", 0, ""},
		{"Error", s.URL, http.StatusBadRequest, "", "search", OperationGet, []string{}, false, false, "", 0, ""},
		{"ErrorMessage", s.URL, http.StatusBadRequest, "blah", "search", OperationGet, []string{}, false, false, "", 0, ""},
		{"BadURL", "http://127.0.0.1:123456", 0, "", "search", OperationGet, []string{}, false, false, "", 0, ""},
		{"InvalidSearch", s.URL, 0, "", "", OperationGet, []string{}, false, false, "", 0, ""},
		{"InvalidOperation", s.URL, 0, "", "search", "", []string{}, false, false, "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code
			m.message = tt.message
			m.search = tt.search
			m.op = tt.op
			m.options = strings.Join(tt.options, ",")
			m.fingerprint = tt.fingerprint
			m.exact = tt.exact
			m.pageToken = tt.pageToken
			if tt.pageSize == 0 {
				m.pageSize = ""
			} else {
				m.pageSize = strconv.Itoa(tt.pageSize)
			}
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
					} else if got, want := err.Message, tt.message; got != want {
						t.Errorf("got message %v, want %v", got, want)
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
		message string
		fp      [20]byte
	}{
		{"Success", s.URL, http.StatusOK, "", fp},
		{"Error", s.URL, http.StatusBadRequest, "", fp},
		{"ErrorMessage", s.URL, http.StatusBadRequest, "blah", fp},
		{"BadURL", "http://127.0.0.1:123456", 0, "", fp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.code = tt.code
			m.message = tt.message

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
					} else if got, want := err.Message, tt.message; got != want {
						t.Errorf("got message %v, want %v", got, want)
					}
				}
			}
		})
	}
}
