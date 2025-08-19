package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
}

func handlerFunc(w *response.Writer, req *request.Request)  {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		writeResponse(w, response.StatusBadRequest)
	case "/myproblem":
		writeResponse(w, response.StatusInternalServerError)
	default:
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
