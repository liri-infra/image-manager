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
	endpoint   string
	userAgent  string
	httpClient *http.Client
	token      string
}

// NewClient creates a new client connecting to the specified endpoint.
func NewClient(endpoint, token string) (*Client, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DisableCompression: false,
	}
	httpClient := &http.Client{Transport: transport, Timeout: 60 * time.Minute}

	return &Client{endpoint, "image-manager", httpClient, token}, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", c.endpoint, path))
	if err != nil {
		return nil, err
	}

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

// UploadSingle uploads file path to channel.
func (c *Client) UploadSingle(channel, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	defer writer.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	part, err := writer.CreateFormFile("file", fileInfo.Name())
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("%s/api/v1/upload/%s", c.endpoint, channel))
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", u.String(), body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.userAgent)
	request.Header.Set("Authorization", fmt.Sprintf("BEARER %s", c.token))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", response.Status)
	}

	return nil
}

// Upload uploads multiple files listed in paths at once to channel.
func (c *Client) Upload(channel string, paths []string) error {
	r, w := io.Pipe()
	writer := multipart.NewWriter(w)

	errChan := make(chan error)

	f := func() {
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
	}

	u, err := url.Parse(fmt.Sprintf("%s/api/v1/upload/%s", c.endpoint, channel))
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", u.String(), r)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.userAgent)
	request.Header.Set("Authorization", fmt.Sprintf("BEARER %s", c.token))

	go f()

	if _, err := c.httpClient.Do(request); err != nil {
		return err
	}

	err = <-errChan
	if err != nil {
		return err
	}

	return nil
}
