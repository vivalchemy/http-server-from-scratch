package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"vivalchemy/http-server-from-scratch/internal/request"
	"vivalchemy/http-server-from-scratch/internal/response"
	"vivalchemy/http-server-from-scratch/internal/server"
)

const port = 42069

func respond200() []byte {
	return []byte(`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
	`)
}

func respond400() []byte {
	return []byte(`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)

}

func respond500() []byte {
	return []byte(`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		status := response.StatusOk
		body := respond200()
		if req.TargetPath == "/yourproblem" && req.Method == "GET" {
			status = response.StatusBadRequest
			body = respond400()
		} else if req.TargetPath == "/myproblem" && req.Method == "GET" {
			status = response.StatusInternalServerError
			body = respond500()
		}

		h.Replace("Content-Type", "text/html")
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteStatusLine(status)
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
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
