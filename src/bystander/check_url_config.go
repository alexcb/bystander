package bystander

import (
	"time"
)

// URLCheckConfig defines a URL check
type URLCheckConfig struct {
	CheckCommonConfig
	url     string
	timeout time.Duration
}

func parseURLCheck(c map[interface{}]interface{}) CheckConfig {
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

func (s *URLCheckConfig) Init(vars map[string]string) (Check, error) {
	c := &URLCheck{}
	initCheckCommon(c, s, vars)

	c.url = subVar(s.url, c.Common().tags)
	c.timeout = s.timeout

	return c, nil
}
