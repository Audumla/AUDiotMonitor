package templatex

import "regexp"

var tokenRE = regexp.MustCompile(`\$\{([^}]+)\}`)

// Expand replaces ${key} tokens in s using resolve.
// If resolve returns ok=false, the token is replaced with an empty string.
func Expand(s string, resolve func(key string) (value string, ok bool)) string {
	if s == "" {
		return s
	}
	return tokenRE.ReplaceAllStringFunc(s, func(token string) string {
		m := tokenRE.FindStringSubmatch(token)
		if len(m) != 2 {
			return token
		}
		if v, ok := resolve(m[1]); ok {
			return v
		}
		return ""
	})
}
