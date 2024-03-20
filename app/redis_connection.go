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

type Request struct {
	Lines    []string
	Commands []Command
}

func (rd *Redis) handleConnection(conn net.Conn) {
	//defer conn.Close()

	fmt.Println("New connection from: ", conn.RemoteAddr().String())

	for {
		reqs, err := rd.buildRequest(conn)
		if err != nil {
			fmt.Println("Error reading data: ", err.Error())
			os.Exit(1)
		}

		go func() {
			if err := rd.handleResponseLines(reqs.Lines, &reqs.Commands); err != nil {
				fmt.Println("Error handleResponseLines: ", err.Error())
				os.Exit(1)
			}

			for com := 0; com < len(reqs.Commands); com++ {
				fmt.Println(
					"Now running:",
					interface{}(reqs.Commands[com].formatCommand()),
				)
				if err := rd.runCommand(reqs.Commands[com], conn); err != nil {
					fmt.Println("Error runCommand:", err.Error())
					os.Exit(1)
				}
			}
		}()

		time.Sleep(30 * time.Millisecond)
	}
}

func (rd *Redis) buildRequest(conn net.Conn) (req Request, err error) {

	readBuffer := make([]byte, 1024)

	n, err := conn.Read(readBuffer)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
	}

	req.Lines = strings.Split(string(readBuffer[:n]), "\r\n")

	for line := 0; line < len(req.Lines); line++ {
		fmt.Println(req.Lines[line])
	}

	return req, nil
}

func (rd *Redis) handleResponseLines(reqLine []string, commands *[]Command) error {
	if commands == nil {
		commands = &[]Command{}
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
				if strings.HasPrefix(reqLine[j], "$") {
					j++
					command = append(command, reqLine[j])
				}
			}
			newCommand := Command{command[0], command[1:]}
			*commands = append(*commands, newCommand)
			fmt.Println("inserted command:", interface{}(newCommand.formatCommand()))
			i += 2*n + 1
		default:
			i++
		}
	}

	return nil
}
