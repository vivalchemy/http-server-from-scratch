package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"vivalchemy/http-server-from-scratch/headers"
	"vivalchemy/http-server-from-scratch/request"
	"vivalchemy/http-server-from-scratch/response"
	"vivalchemy/http-server-from-scratch/server"
)

const port = 5173

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

func yourProblemHander(w *response.Writer, req *request.Request) {
	body := respond400()
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")

	w.WriteStatusLine(response.StatusBadRequest)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func myProblemHandler(w *response.Writer, req *request.Request) {
	body := respond500()
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")

	w.WriteStatusLine(response.StatusInternalServerError)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func videoHandler(w *response.Writer, req *request.Request) {
	f, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		body := respond500()
		h := response.GetDefaultHeaders(len(body))
		h.Replace("Content-Type", "text/html")

		w.WriteStatusLine(response.StatusInternalServerError)
		w.WriteHeaders(*h)
		w.WriteBody(body)
		return
	}

	h := response.GetDefaultHeaders(len(f))
	h.Replace("Content-Type", "video/mp4")
	h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))

	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)
	w.WriteBody(f)

}

func proxyHandler(fullUrl string, domain string) server.Handler {
	return func(w *response.Writer, req *request.Request) {
		joined, _ := url.JoinPath(fullUrl, strings.TrimPrefix(req.TargetPath, domain))
		res, err := http.Get(joined)
		if err != nil {
			body := respond500()
			h := response.GetDefaultHeaders(len(body))
			h.Replace("Content-Type", "text/html")

			w.WriteStatusLine(response.StatusInternalServerError)
			w.WriteHeaders(*h)
			w.WriteBody(body)
			return
		}
		defer res.Body.Close()

		h := response.GetDefaultHeaders(0)
		h.Delete("Content-Length")
		h.Set("Transfer-Encoding", "chunked")
		h.Replace("Content-Type", res.Header.Get("Content-Type"))
		h.Set("Trailers", "X-Content-SHA256, X-Content-Length")

		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(*h)

		fullBody := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := res.Body.Read(buf)
			if n > 0 {
				fullBody = append(fullBody, buf[:n]...)
				w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
				w.WriteBody(buf[:n])
				w.WriteBody([]byte("\r\n"))
			}
			// if err == io.EOF {
			// 	break
			// }
			if err != nil {
				break
			}
		}

		// end of chunks
		w.WriteBody([]byte("0\r\n"))

		// trailers
		trailers := headers.NewHeaders()
		out := sha256.Sum256(fullBody)
		trailers.Set("X-Content-SHA256", toStr(out[:]))
		trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
		w.WriteHeaders(*trailers)
		w.WriteBody([]byte("\r\n"))

	}
}

func allHandler(w *response.Writer, req *request.Request) {
	body := respond200()
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")

	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func main() {
	startTime := time.Now()
	s := server.NewServer()
	// -----------------
	// Simple HTML routes
	// -----------------
	s.Get("/*", allHandler)
	s.Get("/yourproblem", yourProblemHander)
	s.Get("/myproblem", myProblemHandler)

	// -----------------
	// Serve a static video
	// -----------------
	s.Get("/video", videoHandler)

	// -----------------
	// Proxy to httpbin with chunked encoding
	// -----------------
	s.Get("/httpbin/*", proxyHandler("https://httpbin.org/", "/httpbin/"))
	s.Get("/daily/*", proxyHandler("https://daily.dev/", "/daily/"))
	s.Get("/wiki/*", proxyHandler("https://www.wikipedia.org/wiki/", "/wiki/"))
	s.Get("/ddg/*", proxyHandler("https://duckduckgo.com/", "/ddg/"))
	s.Get("/vivalchemy/*", proxyHandler("https://vivalchemy.github.io/", "/vivalchemy/"))

	defer s.Close()
	err := s.Serve(port)
	totalTime := time.Since(startTime)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	fmt.Printf("Server started on port :%d\n", port)
	fmt.Println("Time Taken:", totalTime)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Server gracefully stopped")
}
