package request

import (
	"fmt"
	"io"
	"strings"
	"bytes"
)

type Request struct {
	RequestLine RequestLine
	parserState int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}


const (
	requestStateInitialized int = iota
	requestStateDone
)

const bufferSize int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	req := &Request{
				parserState: requestStateInitialized,
			}
	readToIndex := 0
	for req.parserState != requestStateDone {

		// read into the buffer
		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				req.parserState = requestStateDone
				break
			}
			return nil, fmt.Errorf("error reading: %v\n", err)
		}
		if bytesRead > 0 {
			readToIndex += bytesRead

			// parse from the buffer
			parsedBytes, err := req.parse(buf[:readToIndex])
			if err != nil {
				return nil, fmt.Errorf("error parsing the request-line: %v\n", err)
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

	return requestLine, len(data), nil
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
		return nil, fmt.Errorf("Invalid HTTP version: %s", versionParts[0])
	}
	if versionParts[1] != "1.1" {
		return nil, fmt.Errorf("Invalid HTTP version: %s", versionParts[1])
	}

	return &RequestLine{
		Method: method,
		RequestTarget: target,
		HttpVersion: versionParts[1],
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parserState {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		} else if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.parserState = requestStateDone
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a requestStateDone state")
	default:
		return 0, fmt.Errorf("error unknown state: %v\n", r.parserState)
	}
}
