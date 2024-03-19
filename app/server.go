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

type request struct {
	Lines    []string
	Commands []string
}

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

	for req := 0; req < len(reqs.Lines); req++ {
		fmt.Println("Now handling: " + reqs.Lines[req])
		if err := handleResponseLines(reqs.Lines[req], &reqs.Commands); err != nil {
			fmt.Println("Error handleResponseLines: ", err.Error())
			os.Exit(1)
		}
	}

	for com := 0; com < len(reqs.Commands); com++ {
		fmt.Println("Now running: " + reqs.Commands[com])
		if err := runCommand(reqs.Commands[com], conn); err != nil {
			fmt.Println("Error runCommand:", err.Error())
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

	req.Lines = strings.Split(string(readBuffer[:n]), "\n")

	return req, nil
}

func handleResponseLines(reqLine string, commands *[]string) error {
	lineParts := strings.Split(reqLine, " ")

	if len(lineParts) == 0 {
		return errors.New("no command input")
	}

	if commands == nil {
		commands = &[]string{}
	}

	for i := 0; i < len(lineParts); i++ {
		switch {
		case strings.HasPrefix(lineParts[i], "*") || strings.HasPrefix(lineParts[i], "$") || lineParts[i] == "":
			continue
		default:
			*commands = append(*commands, lineParts[i])
		}
	}

	return nil
}

func runCommand(command string, conn net.Conn) error {

	switch {
	case command == "ping":
		fmt.Println("match \"ping\"")
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			return err
		}
	default:
		return errors.New("no matching command")
	}
	return nil
}
