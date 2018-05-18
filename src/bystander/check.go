package bystander

// Check defines a check
type Check interface {
	// Run runs the check, it returns true if the check is OK, and a map of additional details
	Run() (bool, map[string]string)

	// CommonConfig returns a pointer to the base check config
	CommonConfig() *CheckConfig
}

// CheckConfig contains common config for all check types
type CheckConfig struct {
	tags                      map[string]string
	numFailuresBeforeAlerting int
}

// CommonConfig return the common check config
func (s *CheckConfig) CommonConfig() *CheckConfig {
	return s
}
