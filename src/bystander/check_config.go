package bystander

import (
	"bytes"
	"fmt"
	"unicode"
)

// CheckConfig defines a check configuration
type CheckConfig interface {
	// Run runs the check, it returns true if the check is OK, and a map of additional details
	Init(vars map[string]string) (Check, error)

	// CommonConfig returns a pointer to the base check config
	CommonConfig() *CheckCommonConfig
}

// CheckConfig contains common config for all check types
type CheckCommonConfig struct {
	tags                      map[string]string
	foreach                   map[string]string
	numFailuresBeforeAlerting int
	numSuccessBeforeRecovery  int
	notes                     string
	notifier                  string
}

// CommonConfig return the common check config
func (s *CheckCommonConfig) CommonConfig() *CheckCommonConfig {
	return s
}

func subVar(s string, vars map[string]string) string {
	var buf bytes.Buffer
	var varnameBuf bytes.Buffer
	found := false
	escaped := false
	for _, r := range s {
		if found {
			if varnameBuf.Len() == 0 && r == '{' {
				escaped = true
				continue
			}
			if escaped {
				if r == '}' {
					varname := varnameBuf.String()
					val, ok := vars[varname]
					if !ok {
						panic(fmt.Sprintf("var %q not found", varname))
					}
					buf.WriteString(val)
					varnameBuf.Reset()
					found = false
					escaped = false
					continue
				}
				varnameBuf.WriteRune(r)
				continue
			}
			if unicode.IsLetter(r) {
				varnameBuf.WriteRune(r)
				continue
			}
			varname := varnameBuf.String()
			val, ok := vars[varname]
			if !ok {
				panic(fmt.Sprintf("var %q not found", varname))
			}
			buf.WriteString(val)
			varnameBuf.Reset()
			found = false
			buf.WriteRune(r)
		} else {
			if r == '$' {
				found = true
			} else {
				buf.WriteRune(r)
			}
		}
	}
	return buf.String()
}

func subVars(m, vars map[string]string) map[string]string {
	mm := map[string]string{}
	for k, v := range m {
		mm[k] = subVar(v, vars)
	}
	return mm
}

func mergeVars(m1, m2 map[string]string) map[string]string {
	mm := map[string]string{}
	for k, v := range m1 {
		mm[k] = v
	}
	for k, v := range m2 {
		mm[k] = v
	}
	return mm
}

func initCheckCommon(c Check, cc CheckConfig, vars map[string]string) map[string]string {
	c.Common().tags = subVars(cc.CommonConfig().tags, vars)
	varsAndTags := mergeVars(vars, c.Common().tags)
	c.Common().numFailuresBeforeAlerting = cc.CommonConfig().numFailuresBeforeAlerting
	c.Common().numSuccessBeforeRecovery = cc.CommonConfig().numSuccessBeforeRecovery
	c.Common().notes = subVar(cc.CommonConfig().notes, varsAndTags)
	c.Common().notifier = subVar(cc.CommonConfig().notifier, varsAndTags)
	return varsAndTags
}
