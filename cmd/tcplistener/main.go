package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
	"vivalchemy/http-server-from-scratch/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err.Error(), "Unable to use the port. It maybe occupied")
	}

	fmt.Println("Listening on port :42069")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err.Error(), "Unable to use the port. It maybe occupied")
		}

		startTime := time.Now()
		r, err := request.RequestFromReader(conn)
		parseTime := time.Since(startTime)
		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf("Request Line\n")
		fmt.Printf("- Method %v\n", r.RequestLine.Method)
		fmt.Printf("- Target %v\n", r.RequestLine.TargetPath)
		fmt.Printf("- Version %v\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers Line\n")
		for k, v := range r.Headers.GetAll() {
			fmt.Printf("- %v: %v\n", k, v)
		}
		fmt.Printf("Body:\n")
		fmt.Println(string(r.Body))
		fmt.Println("Time taken to parse: ", parseTime)
	}

}

// DEAD CODE
func getLinesChannel(f io.ReadCloser) <-chan string {
	str := ""
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				break
			}
			buffer = buffer[:n]

			if i := bytes.IndexByte(buffer, '\n'); i != -1 {
				str += string(buffer[:i])

				buffer = buffer[i+1:]
				out <- str
				str = ""

			}
			str += string(buffer)
		}

		if len(str) != 0 {
			out <- str
		}
	}()
	return out
}

func getFilePointerMust(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err.Error(), "Unable to open the string")
	}

	return file
}
