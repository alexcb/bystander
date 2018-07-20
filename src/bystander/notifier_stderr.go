package bystander

import (
	"fmt"
	"os"
)

// StdErrNotifierConfig defines a stderr notifier (useful for debugging bystander)
type StdErrNotifierConfig struct {
	NotifierConfig
}

func parseStdErrNotifier(c map[interface{}]interface{}) Notifier {
	return &StdErrNotifierConfig{}
}

func (s *StdErrNotifierConfig) Init(webAddr string) error {
	return nil
}

// Notify runs the notifier
func (s *StdErrNotifierConfig) Notify(id, checkName string, ok bool, details map[string]string) error {
	os.Stderr.WriteString(fmt.Sprintf("id=%v check=%v ok=%v details=%v\n", id, checkName, ok, details))
	return nil
}
