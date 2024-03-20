package main

import (
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

	fmt.Println("New connection from:", conn.RemoteAddr().String())

	for {
		reqs, err := rd.buildRequest(conn)
		if err != nil {
			fmt.Println("Error reading data:", err.Error())
			conn.Close()
			return
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
		fmt.Println("Error reading data:", err.Error())
		conn.Close()
		return req, err
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

	for i := 0; i < len(reqLine); i++ {
		nextPart := 0
		switch {
		case strings.HasPrefix(reqLine[i], "*"):
			*commands = append(*commands, Command{commandType: "*"})
			if commandLength, err := strconv.Atoi(strings.TrimPrefix(reqLine[i], "*")); err != nil {
				return err
			} else {
				(*commands)[len(*commands)-1].commandLength = commandLength
			}
			continue
		case strings.HasPrefix(reqLine[i], "+"):
			*commands = append(*commands, Command{commandType: "+", command: strings.TrimPrefix(reqLine[i], "+")})
			continue
		case strings.HasPrefix(reqLine[i], "$"):
			if _nextPart, err := strconv.Atoi(strings.TrimPrefix(reqLine[i], "$")); err != nil {
				return err
			} else {
				nextPart = _nextPart
			}
			continue
		default:
			if (*commands)[len(*commands)-1].command == "" {
				(*commands)[len(*commands)-1].command = reqLine[i][:nextPart]
			} else {
				(*commands)[len(*commands)-1].args = append((*commands)[len(*commands)-1].args, reqLine[i][:nextPart])
			}
		}
	}

	return nil
}
