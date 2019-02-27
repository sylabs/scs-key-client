// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	jsonresp "github.com/sylabs/json-resp"
)

// Paths used in this file.
const (
	PathPKSAdd    = "/pks/add"
	PathPKSLookup = "/pks/lookup"
)

// Operations for PKS Add.
const (
	OperationGet    = "get"
	OperationIndex  = "index"
	OperationVIndex = "vindex"
)

// Options for PKS Add.
const (
	OptionMachineReadable = "mr"
)

// PKSAdd submits an ASCII armored keyring to the Key Service, as specified in section 4 of the
// OpenPGP HTTP Keyserver Protocol (HKP) specification. The context controls the lifetime of the
// request.
func (c *Client) PKSAdd(ctx context.Context, keyText string) error {
	v := url.Values{}
	v.Set("keytext", keyText)

	req, err := c.newRequest(http.MethodPost, PathPKSAdd, "", strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return jsonresp.NewError(res.StatusCode, res.Status)
	}
	return nil
}

// PKSLookup requests data from the Key Service, as specified in section 3 of the OpenPGP HTTP
// Keyserver Protocol (HKP) specification. The context controls the lifetime of the request.
func (c *Client) PKSLookup(ctx context.Context, pd *PageDetails, search, operation string, fingerprint, exact bool, options []string) (response string, err error) {
	v := url.Values{}
	v.Set("search", search)
	v.Set("op", operation)
	v.Set("options", strings.Join(options, ","))
	if fingerprint {
		v.Set("fingerprint", "on")
	}
	if exact {
		v.Set("exact", "on")
	}
	if pd != nil {
		v.Set("x-pagesize", strconv.Itoa(pd.size))
		v.Set("x-pagetoken", pd.token)
	}

	req, err := c.newRequest(http.MethodGet, PathPKSLookup, v.Encode(), nil)
	if err != nil {
		return "", err
	}

	res, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", jsonresp.NewError(res.StatusCode, res.Status)
	}

	if pd != nil {
		pd.token = res.Header.Get("X-HKP-Next-Page-Token")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetKey retrieves an ASCII armored keyring from the Key Service. The context controls the
// lifetime of the request.
func (c *Client) GetKey(ctx context.Context, fingerprint [20]byte) (keyText string, err error) {
	return c.PKSLookup(ctx, nil, fmt.Sprintf("%#x", fingerprint), OperationGet, false, true, nil)
}
