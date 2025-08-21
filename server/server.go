package server

import (
	"fmt"
	"io"
	"net"
	"vivalchemy/http-server-from-scratch/request"
	"vivalchemy/http-server-from-scratch/response"
)

type HTTPMethod string

const (
	MethodGet     HTTPMethod = "GET"
	MethodPost    HTTPMethod = "POST"
	MethodDelete  HTTPMethod = "DELETE"
	MethodPut     HTTPMethod = "PUT"
	MethodPatch   HTTPMethod = "PATCH"
	MethodOptions HTTPMethod = "OPTIONS"
	// MethodHead    HTTPMethod = "HEAD"
	// MethodConnect HTTPMethod = "CONNECT"
	// MethodTrace   HTTPMethod = "TRACE"
)

type Server struct {
	closed   bool
	listener net.Listener
	tree     *PathTreeNode
}

func NewServer() *Server {
	return &Server{closed: false, listener: nil, tree: NewPathTree()}
}

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

	// NOTE: read the request path here
	// instead of this use the tree from the server
	handler, err := s.tree.find(HTTPMethod(r.Method), r.TargetPath)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusNotFound)
		responseWriter.WriteHeaders(*headers)
		return
	}

	handler(responseWriter, r)
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

func (s *Server) Serve(port uint16) error {

	s.addOptions() // recursively add the options to each node of the tree
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	go s.run(listener)

	return nil
}

func (s *Server) Close() error {
	s.closed = true
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}

// NOTE: add the methods to the server here
func (s *Server) Get(path string, handler Handler) {
	s.tree.add(MethodGet, path, handler)
}

func (s *Server) Post(path string, handler Handler) {
	s.tree.add(MethodPost, path, handler)
}

func (s *Server) Put(path string, handler Handler) {
	s.tree.add(MethodPut, path, handler)
}

func (s *Server) Delete(path string, handler Handler) {
	s.tree.add(MethodDelete, path, handler)
}

func (s *Server) Patch(path string, handler Handler) {
	s.tree.add(MethodPatch, path, handler)
}

func (s *Server) addOptions() {
	// traverse the tree and add the options handler to all the leaf nodes
	s.tree.addOptions()
}

func (s *Server) AddHandler(method HTTPMethod, path string, handler Handler) {
	s.tree.add(method, path, handler)
}
