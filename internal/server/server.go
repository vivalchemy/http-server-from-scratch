package server

import (
	"fmt"
	"io"
	"net"
	"vivalchemy/http-server-from-scratch/internal/request"
	"vivalchemy/http-server-from-scratch/internal/response"
)

type Server struct {
	closed   bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(res *response.Writer, req *request.Request)

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	headers := response.GetDefaultHeaders(0)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*headers)
		return
	}

	s.handler(responseWriter, r)

}

func (s *Server) run(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		if s.closed {
			return
		}
		go s.handle(conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{closed: false, listener: listener, handler: handler}
	go server.run(listener)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}
