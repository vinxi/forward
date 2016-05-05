package router

import "net/url"

// Match matches the given path string againts the register pattern.
func Match(pat, path string) (url.Values, bool) {
	var i, j int
	p := make(url.Values)

	for i < len(path) {
		switch {
		case j >= len(pat):
			if pat != "/" && len(pat) > 0 && pat[len(pat)-1] == '/' {
				return p, true
			}
			return nil, false
		case pat[j] == ':':
			var name, val string
			var nextc byte
			name, nextc, j = match(pat, isAlnum, j+1)
			val, _, i = match(path, matchPart(nextc), i)
			p.Add(":"+name, val)
		case path[i] == pat[j]:
			i++
			j++
		default:
			return nil, false
		}
	}

	if j != len(pat) {
		return nil, false
	}

	return p, true
}

// Tail returns the trailing string in path after the final slash for a pat ending with a slash.
//
// Examples:
//
//  Tail("/hello/:title/", "/hello/mr/mizerany") == "mizerany"
//  Tail("/:a/", "/x/y/z")                       == "y/z"
//
func Tail(pat, path string) string {
	var i, j int
	for i < len(path) {
		switch {
		case j >= len(pat):
			if pat[len(pat)-1] == '/' {
				return path[i:]
			}
			return ""
		case pat[j] == ':':
			var nextc byte
			_, nextc, j = match(pat, isAlnum, j+1)
			_, _, i = match(path, matchPart(nextc), i)
		case path[i] == pat[j]:
			i++
			j++
		default:
			return ""
		}
	}
	return ""
}

func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

type matcher func(byte) bool

func match(s string, f matcher, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}
	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}
