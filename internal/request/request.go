package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	ParserState int
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	requestStateInitialized int = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

const bufferSize int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	req := &Request{
		ParserState: requestStateInitialized,
		Headers:     headers.Headers{},
	}
	readToIndex := 0
	for req.ParserState != requestStateDone {
		// read into the buffer
		bytesRead, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if err == io.EOF {
				// if EOF is reached while parsing the body, ensure Content-Length has been satisfied
				if req.ParserState == requestStateParsingBody {
					if contentLengthStr, ok := req.Headers["content-length"]; ok {
						contentLengthNumber, err := strconv.Atoi(contentLengthStr)
						if err != nil {
							return nil, fmt.Errorf("invalid Content-Length: %v", err)
						}
						if len(req.Body) < contentLengthNumber {
							return nil, fmt.Errorf("unexpected EOF: body shorter than Content-Length header")
						}
					}
				}
				req.ParserState = requestStateDone
				break
			}
			return nil, fmt.Errorf("error reading: %v", err)
		}
		if bytesRead > 0 {
			readToIndex += bytesRead

			// parse from the buffer
			parsedBytes, err := req.parse(buf[:readToIndex])
			if err != nil {
				return nil, fmt.Errorf("error parsing the request-line: %v", err)
			}
			if parsedBytes > 0 {
				copy(buf, buf[parsedBytes:readToIndex])
				readToIndex -= parsedBytes
			}
			if readToIndex == len(buf) {
				newBuf := make([]byte, 2*len(buf))
				copy(newBuf, buf)
				buf = newBuf
			}
		}
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		// return nil, 0, fmt.Errorf("could not find CLRF in request-line")
		return nil, 0, nil
	}
	requestLineString := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineString)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	rlParts := strings.Split(str, " ")
	if len(rlParts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}
	method := rlParts[0]
	target := rlParts[1]
	version := rlParts[2]

	// check method
	for _, ch := range method {
		if ch < 'A' || ch > 'Z' {
			return nil, fmt.Errorf("invalid method name: %s", method)
		}
	}

	// check HTTP-version
	versionParts := strings.Split(version, "/")
	if versionParts[0] != "HTTP" {
		return nil, fmt.Errorf("invalid HTTP version: %s", versionParts[0])
	}
	if versionParts[1] != "1.1" {
		return nil, fmt.Errorf("invalid HTTP version: %s", versionParts[1])
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   versionParts[1],
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		} else if n == 0 {
			return totalBytesParsed, nil
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		} else if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.ParserState = requestStateParsingHeaders
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a requestStateDone state")
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLength, ok := r.Headers["content-length"]
		if ok {
			contentLengthNumber, err := strconv.Atoi(contentLength)
			if err != nil {
				return 0, err
			}
			r.Body = append(r.Body, data...)
			if len(r.Body) > contentLengthNumber {
				return 0, fmt.Errorf("error: the length of the body is greater than the Content-Length header")
			} else if len(r.Body) == contentLengthNumber {
				r.ParserState = requestStateDone
			}
		} else {
			r.ParserState = requestStateDone
		}
		return len(data), nil
	default:
		return 0, fmt.Errorf("error unknown state: %v", r.ParserState)
	}
}
