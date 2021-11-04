package porto

import (
	"fmt"
	"regexp"
	"strings"
)

// StdExcludeDirRegexps is the standard directory exclusion list from golangci-lint.
// See https://github.com/golangci/golangci-lint/blob/master/pkg/packages/skip.go.
var StdExcludeDirRegexps = []*regexp.Regexp{
	regexp.MustCompile("^vendor$"),
	regexp.MustCompile("^third_party$"),
	regexp.MustCompile("^testdata$"),
	regexp.MustCompile("^examples$"),
	regexp.MustCompile("^Godeps$"),
	regexp.MustCompile("^builtin$"),
}

func GetRegexpList(regexps string) ([]*regexp.Regexp, error) {
	var regexes []*regexp.Regexp
	if len(regexps) > 0 {
		for _, sfrp := range strings.Split(regexps, ",") {
			sfr, err := regexp.Compile(sfrp)
			if err != nil {
				return nil, fmt.Errorf("failed to compile regex %q: %w", sfrp, err)
			}
			regexes = append(regexes, sfr)
		}
	}

	return regexes, nil
}
