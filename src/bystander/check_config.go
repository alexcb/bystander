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
	hide                      []string
	numFailuresBeforeAlerting int
	numSuccessBeforeRecovery  int
	notes                     string
	notifier                  string
}

// CommonConfig return the common check config
func (s *CheckCommonConfig) CommonConfig() *CheckCommonConfig {
	return s
}

func subVar(s string, vars map[string]string, allowMissing bool) string {
	var redacted = "<redacted>"
	var buf bytes.Buffer
	var varnameBuf bytes.Buffer
	found := false
	escaped := false
	for _, r := range s {
		if found {
			if varnameBuf.Len() == 0 && r == '$' {
				buf.WriteRune('$')
				found = false
				continue
			}
			if varnameBuf.Len() == 0 && r == '{' {
				escaped = true
				continue
			}
			if escaped {
				if r == '}' {
					varname := varnameBuf.String()
					val, ok := vars[varname]
					if !ok {
						if allowMissing {
							val = redacted
						} else {
							panic(fmt.Sprintf("var %q not found", varname))
						}
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
			if unicode.IsLetter(r) || r == '_' {
				varnameBuf.WriteRune(r)
				continue
			}
			varname := varnameBuf.String()
			val, ok := vars[varname]
			if !ok {
				if allowMissing {
					val = redacted
				} else {
					panic(fmt.Sprintf("var %q not found", varname))
				}
			}
			buf.WriteString(val)
			varnameBuf.Reset()
			found = false
		}
		if r == '$' {
			found = true
		} else {
			buf.WriteRune(r)
		}

	}
	if found {
		if escaped {
			panic("missing closing variable character \"}\"")
		}
		varname := varnameBuf.String()
		val, ok := vars[varname]
		if !ok {
			if allowMissing {
				val = redacted
			} else {
				panic(fmt.Sprintf("var %q not found", varname))
			}
		}
		buf.WriteString(val)
		varnameBuf.Reset()
		found = false
	}

	return buf.String()
}

func subVars(m, vars map[string]string, allowMissing bool) map[string]string {
	mm := map[string]string{}
	for k, v := range m {
		mm[k] = subVar(v, vars, allowMissing)
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

func removeVars(m map[string]string, hide []string) map[string]string {
	mm := map[string]string{}
	for k, v := range m {
		mm[k] = v
	}
	for _, k := range hide {
		delete(mm, k)
	}
	return mm
}

func initCheckCommon(c Check, cc CheckConfig, vars map[string]string) {
	c.Common().tags = mergeVars(vars, subVars(cc.CommonConfig().tags, vars, false))
	c.Common().tagsPublic = removeVars(c.Common().tags, cc.CommonConfig().hide)
	c.Common().numFailuresBeforeAlerting = cc.CommonConfig().numFailuresBeforeAlerting
	c.Common().numSuccessBeforeRecovery = cc.CommonConfig().numSuccessBeforeRecovery
	c.Common().notes = subVar(cc.CommonConfig().notes, c.Common().tags, false)
	c.Common().notifier = subVar(cc.CommonConfig().notifier, c.Common().tags, false)
}
