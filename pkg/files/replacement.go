package files

import (
	"regexp"
	"strings"
)

type Replacement struct {
	Pattern     string
	Replacement string
	Lower       bool
}

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
