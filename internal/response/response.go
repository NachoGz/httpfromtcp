package response

import (
	"fmt"
	"io"
	"strings"

	"httpfromtcp/internal/headers"

)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusOK:
		w.Write([]byte("HTTP/1.1 200 OK \r\n"))
	case StatusBadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request \r\n"))
	case StatusInternalServerError:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error \r\n"))
	default:
		return fmt.Errorf("error: incorrect status code: %v", statusCode)
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

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		header := fmt.Sprintf("%s: %s", strings.Title(key), value)
		_, err := w.Write([]byte(fmt.Sprintf("%s\r\n", header)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
