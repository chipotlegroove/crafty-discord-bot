// Package requests contains a helper to avoid repeating code for common HTTP requests
package requests

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

var client *http.Client

func Init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = &http.Client{Transport: tr}
}

func ExecuteRequest(method string, url string, auth string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("error on response: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("error while reading response: %w", err)
	}

	return body, nil
}
