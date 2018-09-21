package bystander

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// URLCheckConfig defines a URL check
type URLCheck struct {
	CheckCommon
	url     string
	timeout time.Duration
}

// Command returns the command
func (s *URLCheck) Command() []string {
	return []string{"curl", s.url}
}

// Run runs the check
func (s *URLCheck) Run() (bool, map[string]string) {

	client := http.Client{
		Timeout: s.timeout,
	}
	resp, err := client.Get(s.url)

	if err != nil {
		return false, map[string]string{
			"url": s.url,
			"err": fmt.Sprintf("%v", err),
		}
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, map[string]string{
			"url":    s.url,
			"err":    fmt.Sprintf("%v", err),
			"body":   truncateString(string(responseData), 300),
			"status": fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	ok := resp.StatusCode == http.StatusOK
	return ok, map[string]string{
		"url":    s.url,
		"body":   truncateString(string(responseData), 300),
		"status": fmt.Sprintf("%d", resp.StatusCode),
	}
}
