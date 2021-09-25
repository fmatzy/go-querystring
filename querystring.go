package querystring

// this code is derived from src/net/url/url.go of github.com/golang/go

import (
	"bytes"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/text/transform"
)

const upperhex = "0123456789ABCDEF"

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func Parse(query string, t transform.Transformer) (url.Values, error) {
	m := make(url.Values)
	err := parseQuery(m, query, t)
	return m, err
}

func parseQuery(m url.Values, query string, t transform.Transformer) (err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := QueryUnescape(key, t)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = QueryUnescape(value, t)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return err
}

func QueryUnescape(s string, t transform.Transformer) (string, error) {
	return unescape(s, true, t)
}

func unescape(s string, comp bool, t transform.Transformer) (string, error) {
	// Count %, check that they're well-formed.
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", url.EscapeError(s)
			}
			i += 3
		case '+':
			hasPlus = comp
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	var buf bytes.Buffer
	buf.Grow(len(s) - 2*n)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			buf.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 2
		case '+':
			if comp {
				buf.WriteByte(' ')
			} else {
				buf.WriteByte('+')
			}
		default:
			buf.WriteByte(s[i])
		}
	}
	b, _, err := transform.Bytes(t, buf.Bytes())
	return string(b), err
}

func Encode(query url.Values, t transform.Transformer) (string, error) {
	if query == nil {
		return "", nil
	}
	var buf strings.Builder
	var err error
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := query[k]
		keyEscaped, err1 := QueryEscape(k, t)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		for _, v := range vs {
			valueEscaped, err1 := QueryEscape(v, t)
			if err1 != nil {
				if err == nil {
					err = err1
				}
				continue
			}

			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(valueEscaped)
		}
	}
	return buf.String(), err
}

func QueryEscape(s string, t transform.Transformer) (string, error) {
	return escape(s, true, t)
}

func escape(s string, comp bool, t transform.Transformer) (string, error) {
	var err error
	s, _, err = transform.String(t, s)
	if err != nil {
		return "", err
	}

	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			if c == ' ' && comp {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s, nil
	}

	var buf [64]byte
	var token []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		token = buf[:required]
	} else {
		token = make([]byte, required)
	}

	if hexCount == 0 {
		copy(token, s)
		for i := 0; i < len(s); i++ {
			if s[i] == ' ' {
				token[i] = '+'
			}
		}
		return string(token), nil
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && comp:
			token[j] = '+'
			j++
		case shouldEscape(c):
			token[j] = '%'
			token[j+1] = upperhex[c>>4]
			token[j+2] = upperhex[c&15]
			j += 3
		default:
			token[j] = s[i]
			j++
		}
	}
	return string(token), nil
}

func shouldEscape(c byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false
	}

	// Everything else must be escaped.
	return true
}
