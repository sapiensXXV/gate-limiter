package util

import (
	"log/slog"
	"regexp"
	"sync"
)

var regexCache sync.Map

func MatchPlain(s1 string, s2 string) bool {
	return s1 == s2
}

func MatchRegex(target string, regex string) bool {
	var r *regexp.Regexp

	if cached, ok := regexCache.Load(regex); ok {
		r = cached.(*regexp.Regexp)
	} else {
		var err error
		r, err = regexp.Compile(regex)
		if err != nil {
			slog.Error("regex compile error", "pattern", regex, "error", err)
			return false
		}
	}

	return r.MatchString(target)
}
