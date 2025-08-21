# HTTP Server From Scratch

A simple HTTP/1.1 server implementation written in Go, built entirely from scratch without using the standard `net/http` package. This project demonstrates low-level HTTP protocol handling and serves as an educational tool for understanding how web servers work under the hood.

## ⚠️ Important Disclaimer

**This is NOT a production-ready server.** This implementation is created purely for learning purposes and should never be used in any real-world projects. It lacks many critical features required for production use, including:

- Security hardening
- Performance optimizations
- Complete HTTP/1.1 specification compliance
- Error handling robustness
- Resource management
- Proper connection handling

## Features

- **Trie-based routing**: Efficient path matching using a tree data structure
- **Wildcard support**: Routes with `*` wildcards for flexible path matching
- **Basic HTTP methods**: Support for GET, POST, PUT, DELETE, PATCH, and OPTIONS
- **Custom request/response handling**: Built from scratch without standard library HTTP components

## Not Implemented into the library but example included for how to do them manually
- **Reverse proxy capabilities**: Basic proxying to external services
- **Static file serving**: Serve static assets like videos
- **Chunked transfer encoding**: Support for streaming responses

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Installation

```bash
git clone https://github.com/vivalchemy/http-server-from-scratch.git
cd http-server-from-scratch
go mod tidy
```

### Running the Server

```bash
go run ./cmd/httpserver/main.go
```

The server will start on port `5173` by default.

## Example Routes

The main server includes several example routes:

- `GET /*` - Catch-all route returning a success page
- `GET /yourproblem` - Returns a 400 Bad Request
- `GET /myproblem` - Returns a 500 Internal Server Error
- `GET /video` - Serves a static MP4 file
- `GET /httpbin/*` - Proxies requests to httpbin.org
- `GET /daily/*` - Proxies requests to daily.dev
- `GET /wiki/*` - Proxies requests to Wikipedia
- `GET /ddg/*` - Proxies requests to DuckDuckGo
- `GET /vivalchemy/*` - Proxies requests to vivalchemy.github.io

## Usage Example

```go
package main

import (
    "github.com/vivalchemy/http-server-from-scratch/server"
    "github.com/vivalchemy/http-server-from-scratch/response"
    "github.com/vivalchemy/http-server-from-scratch/request"
)

func main() {
    s := server.NewServer()
    
    // Simple route
    s.Get("/hello", func(w *response.Writer, req *request.Request) {
        body := []byte("Hello, World!")
        h := response.GetDefaultHeaders(len(body))
        h.Replace("Content-Type", "text/plain")
        
        w.WriteStatusLine(response.StatusOk)
        w.WriteHeaders(*h)
        w.WriteBody(body)
    })
    
    // Wildcard route
    s.Get("/api/*", apiHandler)
    
    s.Serve(8080)
}
```

## Architecture

### Request Flow

1. **TCP Connection**: Accept incoming connections
2. **Request Parsing**: Parse HTTP request line, headers, and body
3. **Routing**: Use trie-based router to find matching handler
4. **Handler Execution**: Execute the matched route handler
5. **Response Writing**: Write status line, headers, and body back to client

### Key Components

- **Headers Package**: Manages HTTP header parsing and manipulation
- **Request Package**: Handles HTTP request parsing with streaming support
- **Response Package**: Provides utilities for writing HTTP responses
- **Server Package**: Core server logic with trie-based routing system

## Limitations

- Only implements a subset of HTTP/1.1
- No HTTPS/TLS support
- No keep-alive connections
- Limited error handling
- No request/response middleware
- No authentication or authorization
- Not optimized for performance or memory usage

## Learning Objectives

This project helps understand:

- HTTP protocol fundamentals
- TCP socket programming in Go
- Request/response parsing
- Routing algorithms (trie data structure)
- Concurrent connection handling
- Stream processing
- Trie based routing

## Contributing

This is an educational project, but feel free to submit issues or pull requests if you find bugs or want to add educational value.

## License

MIT License - see [LICENSE] file for details.

---

**Remember: This is for learning purposes only. Do not use in production!**
