// SPDX-FileCopyrightText: 2020 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/liri-infra/image-manager/internal/common"
	"github.com/liri-infra/image-manager/internal/logger"
)

// Client is used to connect to the server.
type Client struct {
	url        *url.URL
	userAgent  string
	httpClient *http.Client
	token      string
}

// NewClient creates a new client connecting to the specified endpoint.
func NewClient(endpoint, token string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DisableCompression: false,
	}
	httpClient := &http.Client{Transport: transport, Timeout: 60 * time.Minute}

	return &Client{u, "image-manager", httpClient, token}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.url.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.userAgent)
	request.Header.Set("Authorization", fmt.Sprintf("BEARER %s", c.token))
	return request, nil
}

func (c *Client) do(request *http.Request, v interface{}) (*http.Response, error) {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Errorf("Cannot parse response: %v", err)
		return response, err
	}

	bodyString := strings.TrimSuffix(string(body), "\n")

	if response.StatusCode != http.StatusOK {
		return response, errors.New(bodyString)
	}

	if v != nil {
		err = json.Unmarshal(body, v)
		if err != nil {
			logger.Errorf("Error decoding response: %v", err)
			if e, ok := err.(*json.SyntaxError); ok {
				logger.Errorf("Syntax error at byte offset %d", e.Offset)
			}
			logger.Infof("Response: %q", body)
			return nil, err
		}
	}

	return response, nil
}

// Upload uploads an object
func (c *Client) Upload(channel string, paths []string) error {
	r, w := io.Pipe()
	writer := multipart.NewWriter(w)

	errChan := make(chan error)

	go func() {
		defer func() {
			writer.Close()
			w.Close()
			errChan <- nil
		}()

		for _, path := range paths {
			// File entry
			part, err := writer.CreateFormFile("file", filepath.Base(path))
			if err != nil {
				errChan <- err
				return
			}

			// Calculate checksum
			checksum, err := common.CalculateChecksum(path)
			if err != nil {
				errChan <- err
				return
			}

			// Open source file
			file, err := os.Open(path)
			if err != nil {
				errChan <- err
				return
			}

			// Upload
			if _, err = io.Copy(part, file); err != nil {
				file.Close()
				errChan <- err
				return
			}

			file.Close()

			// Let the server verify the checksum
			if err := writer.WriteField("checksum", fmt.Sprintf("%s:%s", filepath.Base(path), checksum)); err != nil {
				errChan <- err
				return
			}
		}
	}()

	rel := &url.URL{Path: fmt.Sprintf("/api/v1/upload/%s", channel)}
	u := c.url.ResolveReference(rel)

	request, err := http.NewRequest("PUT", u.String(), r)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.userAgent)
	request.Header.Set("Authorization", fmt.Sprintf("BEARER %s", c.token))

	if _, err := c.httpClient.Do(request); err != nil {
		return err
	}

	err = <-errChan
	if err != nil {
		return err
	}

	return nil
}
