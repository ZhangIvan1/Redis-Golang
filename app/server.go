package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	netType string = "tcp"
	host    string = "0.0.0.0"
	port    string = "6379"
)

type request struct {
	Lines    []string
	Commands [][]string
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

	for {
		reqs, err := buildRequest(conn)
		if err != nil {
			fmt.Println("Error reading data: ", err.Error())
			os.Exit(1)
		}

		go func() {
			if err := handleResponseLines(reqs.Lines, &reqs.Commands); err != nil {
				fmt.Println("Error handleResponseLines: ", err.Error())
				os.Exit(1)
			}

			for com := 0; com < len(reqs.Commands); com++ {
				fmt.Println("Now running: " + formatCommand(reqs.Commands[com]))
				if err := runCommand(reqs.Commands[com], conn); err != nil {
					fmt.Println("Error runCommand:", err.Error())
					os.Exit(1)
				}
			}
		}()

		time.Sleep(30 * time.Millisecond)
	}
}

func buildRequest(conn net.Conn) (req request, err error) {

	readBuffer := make([]byte, 1024)

	n, err := conn.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	req.Lines = strings.Split(string(readBuffer[:n]), "\r\n")

	for line := 0; line < len(req.Lines); line++ {
		fmt.Println(req.Lines[line])
	}

	return req, nil
}

func handleResponseLines(reqLine []string, commands *[][]string) error {
	if commands == nil {
		commands = &[][]string{}
	}

	for i := 0; i < len(reqLine); {
		switch {
		case strings.HasPrefix(reqLine[i], "*"):
			n, err := strconv.Atoi(reqLine[i][1:])
			if err != nil {
				return errors.New("get command parts failed")
			}

			var command []string
			for j := i + 1; j < i+2*n; j++ {
				if strings.HasPrefix(reqLine[i], "$") {
					j++
					command = append(command, reqLine[j])
				}
			}
			*commands = append(*commands, command)
			fmt.Println("inserted command:", formatCommand(command))
			i += len(command)
		default:
			i++
		}
	}

	return nil
}

func runCommand(command []string, conn net.Conn) error {
	switch {
	case strings.HasPrefix(command[0], "ping"):
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			return err
		}
	case strings.HasPrefix(command[0], "echo"):
		formatString := formatCommand(command[0:])
		if _, err := conn.Write([]byte("$" + strconv.Itoa(len(formatString)) + "\r\n" + formatString + "\r\n")); err != nil {
			return err
		}
	default:
		return errors.New("no matching command")
		//return nil
	}
	return nil
}

func formatCommand(command []string) string {
	if len(command) == 0 {
		return "the command is empty!"
	}
	res := ""
	for i := 0; i < len(command); i++ {
		res += command[i]
		res += " "
	}
	return res[:len(res)-1]
}
