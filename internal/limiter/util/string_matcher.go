package util

import (
	"log"
	"regexp"
)

func MatchPlain(s1 string, s2 string) bool {
	return s1 == s2
}

func MatchRegex(target string, regex string) bool {
	r, err := regexp.Compile(regex)
	if err != nil {
		log.Println("error while compile regex:", err)
	}
	return r.MatchString(target)
}
