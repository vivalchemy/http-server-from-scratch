package response

import (
	"fmt"
	"io"
	"vivalchemy/http-server-from-scratch/headers"
)

type Response struct {
}

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
	StatusNotFound            StatusCode = 404
)

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	case StatusNotFound:
		statusLine = []byte("HTTP/1.1 404 Not Found\r\n")
	default:
		return fmt.Errorf("unrecognized error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	var err error = nil
	headerStr := []byte{}
	for k, v := range h.GetAll() {
		headerStr = fmt.Appendf(headerStr, "%s: %s\r\n", k, v)
	}
	headerStr = fmt.Append(headerStr, "\r\n")
	_, err = w.writer.Write(headerStr)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.writer.Write(p)
}
