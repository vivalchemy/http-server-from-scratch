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
	"vivalchemy/http-server-from-scratch/internal/headers"
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

func toStr(b []byte) string {
	out := strings.Builder{}
	for _, c := range b {
		out.WriteString(fmt.Sprintf("%02x", c))
	}
	return out.String()
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
		} else if req.TargetPath == "/video" && req.Method == "GET" {
			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				h.Replace("Content-Type", "video/mp4")
				h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))
				w.WriteStatusLine(status)
				w.WriteHeaders(*h)
				w.WriteBody(f)
				return
			}
		} else if strings.HasPrefix(req.TargetPath, "/httpbin/") && req.Method == "GET" {
			target := req.TargetPath
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOk)
				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Replace("Content-Type", res.Header.Get("Content-Type"))
				h.Set("Trailers", "X-Content-SHA256")
				h.Set("Trailers", "X-Content-Length")
				w.WriteHeaders(*h)

				fullBody := make([]byte, 0)
				for {
					data := make([]byte, 1024)
					n, err := res.Body.Read(data)
					if err != nil {
						break
						// IDEALLY this should be done and something else for the other errors
						// if errors.Is(io.EOF, err) {
						// 	break
						// }
					}
					fullBody = append(fullBody, data[:n]...)
					w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))

				}
				w.WriteBody([]byte("0\r\n"))
				trailers := headers.NewHeaders()
				out := sha256.Sum256(fullBody)
				trailers.Set("X-Content-SHA256", toStr(out[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailers)
				w.WriteBody([]byte("\r\n"))
				return
			}
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
