package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (rd *Redis) runCommand(command []string, conn net.Conn) error {
	switch {
	case strings.HasPrefix(command[0], "info"):
		info := rd.info()
		fmt.Sprintln("info:", info)
		if _, err := conn.Write([]byte(info)); err != nil {
			return err
		}
	case strings.HasPrefix(command[0], "ping"):
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			return err
		}
	case strings.HasPrefix(command[0], "echo"):
		formatString := rd.formatCommand(command[1:])
		if _, err := conn.Write([]byte("$" + strconv.Itoa(len(formatString)) + "\r\n" + formatString + "\r\n")); err != nil {
			return err
		}
	case strings.HasPrefix(command[0], "set"):
		if err := rd.setStore(command); err != nil {
			return err
		} else {
			if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
				return err
			}
		}
	case strings.HasPrefix(command[0], "get"):
		if length, value, err := rd.getStore(command[1]); err != nil {
			return err
		} else {
			if _, err := conn.Write([]byte(length + value)); err != nil {
				return err
			}
		}
	default:
		return errors.New("no matching command")
		//return nil
	}
	return nil
}

func (rd *Redis) formatCommand(command []string) string {
	if len(command) == 0 {
		return "-------the command is empty!--------"
	}
	res := ""
	for i := 0; i < len(command); i++ {
		res += command[i]
		res += " "
	}
	return res[:len(res)-1]
}
