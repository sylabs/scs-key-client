// Copyright (c) 2019-2020, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	jsonresp "github.com/sylabs/json-resp"
)

const pathVersion = "version"

// VersionInfo contains version information.
type VersionInfo struct {
	Version string `json:"version"`
}

// GetVersion gets version information from the Key Service. The context controls the lifetime of
// the request.
//
// If an non-200 HTTP status code is received, an error wrapping an HTTPError is returned.
func (c *Client) GetVersion(ctx context.Context) (vi VersionInfo, err error) {
	ref := &url.URL{Path: pathVersion}

	req, err := c.NewRequest(ctx, http.MethodGet, ref, nil)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("%w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("%w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return VersionInfo{}, fmt.Errorf("%w", errorFromResponse(res))
	}

	if err := jsonresp.ReadResponse(res.Body, &vi); err != nil {
		return VersionInfo{}, fmt.Errorf("%w", err)
	}
	return vi, nil
}
