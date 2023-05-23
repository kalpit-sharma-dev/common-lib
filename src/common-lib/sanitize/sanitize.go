package sanitize

import (
	"strings"

	"github.com/kennygrant/sanitize"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/validate/is"
)

// HTML strips html tags with a very simple parser, replace common entities, and escape < and > in the result.
// The result is intended to be used as plain text.
func HTML(s string) string {
	return sanitize.HTML(s)
}

// HTMLAllowing parses html and allow certain tags and attributes from the lists
// optionally specified by args - args[0] is a list of allowed tags, args[1] is a list of allowed attributes.
// If either is missing default sets are used.
func HTMLAllowing(s string, args ...[]string) (string, error) {
	return sanitize.HTMLAllowing(s)
}

// Name makes a string safe to use in as name,
// producing a sanitized name replacing . or / with -
// No attempt is made to normalise a path or normalise case.
func Name(s string) string {
	return sanitize.BaseName(s)
}

// FileName makes a string safe to use in a file name,
// producing a sanitized basename replacing . or / with -
// then replacing non-ascii characters.
func FileName(s string) string {
	return sanitize.Name(s)
}

// Identifier makes a string safe to use in as metric name,
// producing a sanitized name replacing ., - or / with _
// No attempt is made to normalise a path or normalise case.
func Identifier(s string) string {
	// Instead of using Regx we are using mapping function as this is quite faster
	// BenchmarkRegx   	   			10000	    170515 ns/op	    4025 B/op	     159 allocs/op
	// BenchmarkMappingFunction   	30000	     40784 ns/op	    2496 B/op	      34 allocs/op
	// https://medium.com/codezillas/golang-replace-vs-regexp-de4e48482f53
	identifierMappingFun := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		}
		return '_'
	}

	s = strings.Map(identifierMappingFun, s)

	fieldFunc := func(c rune) bool {
		return c == '_'
	}

	data := strings.FieldsFunc(s, fieldFunc)

	if len(data) == 0 {
		return ""
	}

	s = strings.Join(data, "_")

	if !is.Identifier(s) {
		return "_" + s
	}
	return s
}

// Path makes a string safe to use as an url path.
func Path(s string) string {
	return sanitize.Path(s)
}

// Number makes a string safe to use as a number or decimal.
func Number(s string) string {
	// Instead of using Regx we are using mapping function as this is quite faster
	// BenchmarkRegx   	   			100000	     18940 ns/op	     630 B/op	      40 allocs/op
	// BenchmarkMappingFunction   	300000	      5774 ns/op	     496 B/op	      18 allocs/op
	// https://medium.com/codezillas/golang-replace-vs-regexp-de4e48482f53
	identifierMappingFun := func(r rune) rune {
		switch {
		case r >= '0' && r <= '9':
			return r
		case r == '.':
			return r
		}
		return -1
	}
	s = strings.Map(identifierMappingFun, s)
	fieldFunc := func(c rune) bool {
		return c == '.'
	}

	data := strings.FieldsFunc(s, fieldFunc)

	if len(data) > 1 {
		return strings.Join(data[:2], ".")
	}

	if len(data) > 0 {
		return data[0]
	}

	return ""
}
