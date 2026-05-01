package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		fmt.Println("Shutting down server..")

		if err := l.Close(); err != nil {
			fmt.Println("Error closing listener:", err)
		}

		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, os.Args)
	}
}

func handleConnection(conn net.Conn, args []string) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	input := string(buf[:n])

	lines := strings.Split(input, "\r\n")
	if len(lines) == 0 {
		return
	}

	reqParts := strings.Split(lines[0], " ")
	if len(reqParts) < 2 {
		return
	}

	headers := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			break
		}
		parts := strings.SplitN(lines[i], ":", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}
	path := reqParts[1]

	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
			fmt.Println("Error writing response:", err)
		}
		return
	}

	switch parts[0] {
	case "echo":
		if len(parts) > 1 {
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(parts[1]), parts[1])
			if _, err := conn.Write([]byte(response)); err != nil {
				fmt.Println("Error writing response:", err)
			}
		} else {
			if _, err := conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n")); err != nil {
				fmt.Println("Error writing response:", err)
			}
		}
	case "user-agent":
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(headers["user-agent"]), headers["user-agent"])
		if _, err := conn.Write([]byte(response)); err != nil {
			fmt.Println("Error writing response:", err)
		}
	case "files":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: your_program --directory path/to/directory\n")
			os.Exit(1)
		}

		if args[1] == "--directory" {

			fullPath := filepath.Join(args[2], parts[1])

			if _, err := os.Stat(fullPath); err != nil {
				if _, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
					fmt.Println("Error writing response:", err)
				}
				return
			}

			file, err := os.Open(fullPath)
			if err != nil {
				fmt.Println("Error opening fil: %w", err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					fmt.Println("Error closing file:", err)
				}
			}()

			content, _ := os.ReadFile(fullPath)

			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
			if _, err := conn.Write([]byte(response)); err != nil {
				fmt.Println("Error writing response:", err)
			}
		} else {
			fmt.Println("missing --directory flag")
			return
		}
	default:
		if _, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n")); err != nil {
			fmt.Println("Error writing response:", err)
		}
	}
}
