package bystander

// CheckConfig defines a check configuration
type Check interface {
	// Run runs the check, it returns true if the check is OK, and a map of additional details
	Run() (bool, map[string]string)

	// CommonConfig returns a pointer to the base check config
	Common() *CheckCommon
}

// CheckConfig contains common config for all check types
type CheckCommon struct {
	tags                      map[string]string
	numFailuresBeforeAlerting int
	numSuccessBeforeRecovery  int
	notes                     string
	notifier                  string
}

// CommonConfig return the common check config
func (s *CheckCommon) Common() *CheckCommon {
	return s
}
