package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Lines    []string
	Commands []Command
}

func (rd *Redis) handleConnection(conn net.Conn) {

	fmt.Println("New connection from:", conn.RemoteAddr().String())

	for {
		reqs, err := rd.buildRequest(conn)
		if err != nil {
			fmt.Println("Error reading data:", err.Error())
			conn.Close()
			return
		}

		if err := rd.handleResponseLines(reqs.Lines, &reqs.Commands); err != nil {
			fmt.Println("Error handleResponseLines: ", err.Error())
			conn.Close()
			return
		}

		for com := 0; com < len(reqs.Commands); com++ {
			fmt.Println(
				"Now running:",
				interface{}(reqs.Commands[com].formatCommand()),
			)
			reqs.Commands[com].commandOffset = len(reqs.Commands[com].buildRequest())
			if err := rd.runCommand(reqs.Commands[com], conn); err != nil {
				fmt.Println("Error runCommand:", err.Error())
				conn.Close()
				return
			}
		}
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
	fmt.Println(req.Lines)

	for line := 0; line < len(req.Lines); line++ {
		fmt.Println(req.Lines[line])
	}

	req.Lines = req.Lines[:len(req.Lines)-1]
	return req, nil
}

func (rd *Redis) handleResponseLines(reqLine []string, commands *[]Command) error {
	if commands == nil {
		commands = &[]Command{}
	}

	nextPart := 0

	for i := 0; i < len(reqLine); i++ {
		fmt.Println("now:", reqLine[i])
		switch {
		case reqLine[i] == "":
		case strings.HasPrefix(reqLine[i], "*") && len(reqLine[i]) > 1:
			*commands = append(*commands, Command{commandType: "*"})
			if commandLength, err := strconv.Atoi(strings.TrimPrefix(reqLine[i], "*")); err != nil {
				return err
			} else {
				(*commands)[len(*commands)-1].commandLength = commandLength
			}
		case strings.HasPrefix(reqLine[i], "+"):
			responseLine := strings.Split(strings.TrimPrefix(reqLine[i], "+"), " ")
			*commands = append(*commands, Command{commandType: "+", command: responseLine[0], args: responseLine[1:]})
		case strings.HasPrefix(reqLine[i], "$"):
			if _nextPart, err := strconv.Atoi(strings.TrimPrefix(reqLine[i], "$")); err != nil {
				return err
			} else {
				nextPart = _nextPart
			}
		default:
			if (*commands)[len(*commands)-1].command == "" {
				(*commands)[len(*commands)-1].command = reqLine[i][:nextPart]
			} else {
				if len(reqLine[i]) > nextPart {
					(*commands)[len(*commands)-1].args = append((*commands)[len(*commands)-1].args, reqLine[i][:nextPart])
					reqLine[i] = reqLine[i][nextPart:]
					i--
				} else {
					(*commands)[len(*commands)-1].args = append((*commands)[len(*commands)-1].args, reqLine[i])
					nextPart -= len(reqLine[i])
				}
			}
		}
	}
	return nil
}

func (rd *Redis) handleConnectionTicker() {
	for {
		if connection, err := rd.listener.Accept(); err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		} else {
			go rd.handleConnection(connection)

		}
	}
}
