package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Command struct {
	commandType   string
	commandLength int
	command       string
	args          []string
}

func (cm *Command) formatCommand() string {
	res := cm.command + " " + cm.formatArgs()
	return res
}

func (cm *Command) formatArgs() string {
	if len(cm.args) == 0 {
		return ""
	}

	res := ""
	for i := 0; i < len(cm.args); i++ {
		res += cm.args[i]
		res += " "
	}
	return res[:len(res)-1]
}

func (rd *Redis) runCommand(command Command, conn net.Conn) error {
	switch {
	case command.commandType == "+":
		if err := rd.handleRepose(command, conn); err != nil {
			return err
		}

	case command.command == "info" || command.command == "INFO":
		info := rd.info()
		if _, err := conn.Write([]byte(info)); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "ping") || strings.HasPrefix(command.command, "PING"):
		if err := rd.handlePing(command, conn); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "echo") || strings.HasPrefix(command.command, "ECHO"):
		args := command.formatArgs()
		if _, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args), args))); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "set") || strings.HasPrefix(command.command, "SET"):
		if err := rd.setStore(command); err != nil {
			return err
		} else if rd.role == MASTER {
			if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
				return err
			}
			if err := rd.doReplication(command); err != nil {
				return err
			}
		}

	case strings.HasPrefix(command.command, "get") || strings.HasPrefix(command.command, "GET"):
		if length, value, err := rd.getStore(command); err != nil {
			return err
		} else if length == -1 {
			if _, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n", length))); err != nil {
				return err
			}
		} else {
			if _, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", length, value))); err != nil {
				return err
			}
		}

	case strings.HasPrefix(command.command, "REPLCONF"):
		if err := rd.handleReplConf(command, conn); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "PSYNC"):
		if err := rd.handlePSync(command, conn); err != nil {
			return err
		}

	default:
		return errors.New("no matching command")
		//return nil
	}
	return nil
}

func (rd *Redis) handleRepose(command Command, conn net.Conn) error {
	switch {

	}
	return nil
}
