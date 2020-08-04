package files

import (
	"fmt"
	"regexp"
	"strings"
)

// Replacement is a regex pattern replacement
type Replacement struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
	Lower       bool   `yaml:"lower"`
}

func (r Replacement) String() string {
	return fmt.Sprintf("'%s' => '%s'", r.Pattern, r.Replacement)
}

// Replace will perform a replacement
func (r Replacement) Replace(path string) (result string, err error) {
	result = path
	pat, err := regexp.Compile(r.Pattern)
	if err != nil {
		return result, err
	}
	if r.Lower {
		return pat.ReplaceAllStringFunc(result, func(s string) string {
			return strings.ToLower(pat.ReplaceAllString(s, r.Replacement))
		}), nil
	}
	return pat.ReplaceAllString(result, r.Replacement), nil
}

// DoReplacements return the result of a series of replacements
func DoReplacements(patterns []Replacement, fullpath string) (result string, err error) {
	result = fullpath
	for _, rep := range patterns {
		result, err = rep.Replace(result)
		if err != nil {
			return
		}
	}
	return
}
