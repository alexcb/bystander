package bystander

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// URLCheckConfig defines a URL check
type URLCheckConfig struct {
	CheckConfig
	url     string
	timeout time.Duration
}

func parseURLCheck(c map[interface{}]interface{}) Check {
	url, ok := c["url"]
	if !ok {
		panic("url missing")
	}

	timeoutStr, ok := c["timeout"]
	if !ok {
		timeoutStr = "1s"
	}
	timeout, err := time.ParseDuration(timeoutStr.(string))
	if err != nil {
		timeout = time.Second
	}

	return &URLCheckConfig{
		url:     url.(string),
		timeout: timeout,
	}
}

// Run runs the check
func (s *URLCheckConfig) Run() (bool, map[string]string) {

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
