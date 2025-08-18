package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	headers := response.GetDefaultHeaders(0)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, req)
	if handlerError != nil {
		errorMessage := handlerError.Message
		headers := response.GetDefaultHeaders(len(errorMessage))

		// write status line and headers to the connection
		response.WriteStatusLine(conn, handlerError.StatusCode)
		response.WriteHeaders(conn, headers)

		// write error message
		conn.Write([]byte(errorMessage))
	} else {
		body := writer.Bytes()
		headers := response.GetDefaultHeaders(len(body))

		// write status line and headers to the connection
		response.WriteStatusLine(conn, response.StatusOK)
		response.WriteHeaders(conn, headers)

		// write the response body from the handler's buffer to the connection
		conn.Write(body)
	}
}


