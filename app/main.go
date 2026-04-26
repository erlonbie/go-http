package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
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

	conn, err := l.Accept()
	defer conn.Close()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	response := "HTTP/1.1 200 OK\r\n\r\n"
	conn.Write([]byte(response))
}
