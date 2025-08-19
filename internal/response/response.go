package response

import (
	"fmt"
	"io"
	"log"
	"strings"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.writer.Write([]byte("HTTP/1.1 200 OK \r\n"))
	case StatusBadRequest:
		w.writer.Write([]byte("HTTP/1.1 400 Bad Request \r\n"))
	case StatusInternalServerError:
		w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error \r\n"))
	default:
		err := fmt.Errorf("error: unrecognized status code: %v", statusCode)
		log.Println(err)
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["content-length"] = fmt.Sprintf("%d", contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	data := []byte{}
	for key, value := range headers {
		data = fmt.Appendf(data, "%s: %s\r\n", strings.Title(key), value)
	}
	data = fmt.Appendf(data, "\r\n")

	_, err := w.writer.Write(data)
	if err != nil {
		log.Printf("error writing headers: %v", err)
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		log.Printf("error writing body: %v", err)
		return 0, err
	}
	return n, nil
}
