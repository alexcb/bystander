package bystander

// CheckConfig defines a check configuration
type Check interface {
	// Run runs the check, it returns true if the check is OK, and a map of additional details
	Run() (bool, map[string]string)

	// Command returns the command to run, this might contain secret hidden values, so it should never be logged or displayed to users
	Command() []string

	// CommandPublic returns a public version of the command that does not contain any hidden (secret) variable values
	CommandPublic() []string

	// CommonConfig returns a pointer to the base check config
	Common() *CheckCommon
}

// CheckConfig contains common config for all check types
type CheckCommon struct {
	tags                      map[string]string
	tagsPublic                map[string]string
	numFailuresBeforeAlerting int
	numSuccessBeforeRecovery  int
	notes                     string
	notifier                  string
}

// CommonConfig return the common check config
func (s *CheckCommon) Common() *CheckCommon {
	return s
}
