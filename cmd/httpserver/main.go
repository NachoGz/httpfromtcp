package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func formatResponse(statusCode response.StatusCode) string {
	body := "<html>\n\t<head>\n"
	body += "\t</head>\n<body>"
	switch statusCode {
	case response.StatusOK:
		body += "\t<title>200 OK</title>\n"
		body += "\t\t<h1>Success!</h1>\n"
		body += "\t\t<p>Your request was an absolute banger.</p>\n"
	case response.StatusBadRequest:
		body += "\t<title>400 Bad Request</title>\n"
		body += "\t\t<h1>Bad Request</h1>\n"
		body += "\t\t<p>Your request honestly kinda sucked.</p>\n"
	case response.StatusInternalServerError:
		body += "\t<title>500 Internal Server Error</title>\n"
		body += "\t\t<h1>Internal Server Error</h1>\n"
		body += "\t\t<p>Okay, you know what? This one is on me.</p>\n"
	}

	body += "\t</body>\n"
	body += "</html>\n"
	return body
}

func writeResponse(w *response.Writer, statusCode response.StatusCode) {
	body := formatResponse(statusCode)

	// write status line and headers to the connection
	w.WriteStatusLine(statusCode)
	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	w.WriteHeaders(h)
	// write the response body from the handler's buffer to the connection
	w.WriteBody([]byte(body))

	w.WriteBody([]byte("\r\n"))
}

func handlerFunc(w *response.Writer, req *request.Request)  {
	if req.RequestLine.RequestTarget == "/yourproblem"{
		writeResponse(w, response.StatusBadRequest)
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		writeResponse(w, response.StatusInternalServerError)
	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		target := req.RequestLine.RequestTarget
		res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
		if err != nil {
			writeResponse(w, response.StatusInternalServerError)
		} else {
			w.WriteStatusLine(response.StatusOK)

			h := response.GetDefaultHeaders(int(res.ContentLength))
			h.Delete("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")
			h.Replace("Content-Type", "text/plain")
			w.WriteHeaders(h)
			fullBody :=[]byte{}
			for {
				data := make([]byte, 32)
				n, err := res.Body.Read(data)
				if err != nil {
					break
				}
				fullBody = append(fullBody, data[:n]...)
				w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
				w.WriteBody(data[:n])
				w.WriteBody([]byte("\r\n"))
			}
			w.WriteBody([]byte("0\r\n"))
			hash := sha256.Sum256(fullBody)
			trailers := headers.NewHeaders()
			trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			w.WriteHeaders(trailers)
			w.WriteBody([]byte("\r\n"))
			return
		}
	} else if req.RequestLine.RequestTarget == "/video" {
		w.WriteStatusLine(response.StatusOK)
		filepath := "assets/vim.mp4"
		video, err := os.ReadFile(filepath)
		if err != nil {
			log.Printf("error reading file: %v", filepath)
		}

		h := response.GetDefaultHeaders(len(video))
		h.Replace("Content-Type", "video/mp4")
		w.WriteHeaders(h)
		// write the response body from the handler's buffer to the connection
		w.WriteBody(video)

		w.WriteBody([]byte("\r\n"))

	} else {
		writeResponse(w, response.StatusOK)
	}
}

func main() {
	server, err := server.Serve(port, handlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan


	log.Println("Server gracefully stopped")
}
