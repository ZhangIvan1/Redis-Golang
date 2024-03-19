package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	netType string = "tcp"
	host    string = "0.0.0.0"
	port    string = "6379"
)

type request []string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen(netType, host+":"+port)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("New connection from: ", conn.RemoteAddr().String())

	reqs, err := buildRequest(conn)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	fmt.Println(reqs)

	for req := 0; req < len(reqs); req++ {
		if err := handleResponse(conn, reqs[req]); err != nil {
			fmt.Println("Error writing output: ", err.Error())
			os.Exit(1)
		}
	}

}

func buildRequest(conn net.Conn) (req request, err error) {
	defer conn.Close()
	readBuffer := make([]byte, 1024)

	n, err := conn.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	req = strings.Split(string(readBuffer[:n]), "\n")

	return req, nil
}

func handleResponse(conn net.Conn, req string) error {
	switch {
	case req == "ping":
		if _, err := conn.Write([]byte("PONG\r\n")); err != nil {
			return err
		}
	default:
		return errors.New("No matching command")
	}
	return nil
}
