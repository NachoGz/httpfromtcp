package server

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	started atomic.Bool // false: not started, true: started
	listener net.Listener
	handler Handler
}

func NewServer(listener net.Listener, started bool, handler Handler) *Server {
	server := &Server{
		listener: listener,
		handler: handler,
	}

	server.started.Store(started)
	return server
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := NewServer(listener, true, handler)

	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	// Set started to false to signal shutdown
	s.started.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	for s.started.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.started.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn)  {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	headers := response.GetDefaultHeaders(0)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(headers)
		return
	}

	s.handler(responseWriter, req)
}


