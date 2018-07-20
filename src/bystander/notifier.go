package bystander

// Notifier defines a notifier
type Notifier interface {
	// Run runs the check, it returns true if the check is OK, and a map of additional details
	Notify(id, checkName string, ok bool, details map[string]string) error

	// Init initializes the notifier
	Init(webAddr, serverID string) error

	// CommonConfig returns a pointer to the base check config
	CommonConfig() *NotifierConfig
}

// NotifierConfig contains common config for all notifiers
type NotifierConfig struct {
	name string
}

// CommonConfig return the common check config
func (s *NotifierConfig) CommonConfig() *NotifierConfig {
	return s
}
