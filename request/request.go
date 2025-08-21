package request

import (
	"bytes"
	"fmt"
	"io"
	"vivalchemy/http-server-from-scratch/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
)

type RequestLine struct {
	Method      string
	TargetPath  string
	HttpVersion string
}

type Request struct {
	RequestLine
	Headers *headers.Headers
	Body    []byte
	state   parserState
}

func NewRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http verison. only HTTP/1.1 support is available")
var ErrorRequestInErrorState = fmt.Errorf("request is in error state")
var SEPERATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:      string(parts[0]),
		TargetPath:  string(parts[1]),
		HttpVersion: string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool {
	return r.Headers.GetIntMust("content-length", 0) > 0
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, doneParsingHeaders, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if doneParsingHeaders {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			contentLength := r.Headers.GetIntMust("content-length", 0)
			if contentLength == 0 {
				r.state = StateDone
				continue
			}

			remainingToRead := min(contentLength-len(r.Body), len(currentData))
			r.Body = append(r.Body, currentData[:remainingToRead]...)
			read += remainingToRead

			if len(r.Body) == contentLength {
				r.state = StateDone
			}

		case StateDone:
			break outer
		default:
			panic("invalid state")
		}
	}
	return read, nil
}

func (r *Request) isDone() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()

	// NOTE: the buffer could overrun a header that exceeds 1k could do that
	// or the body
	buf := make([]byte, 1000)
	bufLen := 0
	for !request.isDone() {
		n, err := reader.Read(buf[bufLen:])
		// TODO: what to do here?
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])

		bufLen -= readN
	}

	return request, nil

}
