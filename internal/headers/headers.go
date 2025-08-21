package headers

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var rn = []byte("\r\n")
var ErrorMalformedHeader = fmt.Errorf("malformed header")
var ErrorMalformedFieldName = fmt.Errorf("malformed field name")

func isToken(s []byte) bool {
	allowed := []byte("!#$%&'*+-.^_`|~")
	for _, ch := range s {
		if !(ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' ||
			bytes.IndexByte(allowed, ch) >= 0) {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	// since the field value can contain the colon
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrorMalformedHeader
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ErrorMalformedFieldName
	}

	return string(name), string(value), nil
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Replace(name string, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

func (h *Headers) Set(name string, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		fmt.Println("Appending header", name)
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) Get(name string) (string, bool) {
	value, ok := h.headers[strings.ToLower(name)]
	return value, ok
}

func (h *Headers) GetAll() map[string]string {
	return h.headers
}

func (h *Headers) GetIntMust(name string, defaultValue int) int {
	valueStr, exists := h.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		// empty header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		if !isToken([]byte(name)) {
			return 0, false, ErrorMalformedFieldName
		}
		read += idx + len(rn)
		h.Set(name, value)
	}

	return read, done, nil
}
