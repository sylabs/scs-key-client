// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Config contains the client configuration.
type Config struct {
	// BaseURL of the service (https://keys.sylabs.io is used if not supplied).
	BaseURL string
	// Auth token to include in the Authorization header of each request (if supplied).
	AuthToken string
	// User agent to include in each request (if supplied).
	UserAgent string
	// HTTPClient to use to make HTTP requests (if supplied).
	HTTPClient *http.Client
}

// DefaultConfig is a configuration that uses default values.
var DefaultConfig = &Config{}

// PageDetails includes pagination details.
type PageDetails struct {
	// Maximum number of results per page (server may ignore or return fewer).
	Size int
	// Token for next page (advanced with each request, empty for last page).
	Token string
}

// Client describes the client details.
type Client struct {
	baseURL    *url.URL
	authToken  string
	userAgent  string
	httpClient *http.Client
}

// NewClient sets up a new Key Service client with the specified base URL and auth token.
func NewClient(cfg *Config) (c *Client, err error) {
	if cfg == nil {
		cfg = DefaultConfig
	}

	// Determine base URL
	bu := "https://keys.sylabs.io"
	if cfg.BaseURL != "" {
		bu = cfg.BaseURL
	}
	baseURL, err := url.Parse(bu)
	if err != nil {
		return nil, err
	}

	c = &Client{
		baseURL:   baseURL,
		authToken: cfg.AuthToken,
		userAgent: cfg.UserAgent,
	}

	// Set HTTP client
	if cfg.HTTPClient != nil {
		c.httpClient = cfg.HTTPClient
	} else {
		c.httpClient = http.DefaultClient
	}

	return c, nil
}

// newRequest returns a new Request given a method, path, query, and optional body.
func (c *Client) newRequest(method, path, rawQuery string, body io.Reader) (r *http.Request, err error) {
	u := c.baseURL.ResolveReference(&url.URL{
		Path:     path,
		RawQuery: rawQuery,
	})

	r, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if v := c.authToken; v != "" {
		r.Header.Set("Authorization", fmt.Sprintf("BEARER %s", v))
	}
	if v := c.userAgent; v != "" {
		r.Header.Set("User-Agent", v)
	}

	return r, nil
}
