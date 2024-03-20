package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Command struct {
	command string
	args    []string
}

func (cm *Command) formatCommand() (int, string) {
	_, args := cm.formatArgs()
	res := cm.command + " " + args
	return len(res), res
}

func (cm *Command) formatArgs() (int, string) {
	res := ""
	for i := 0; i < len(cm.args); i++ {
		res += cm.args[i]
		res += " "
	}
	return len(res) - 1, res[:len(res)-1]
}

func (rd *Redis) runCommand(command Command, conn net.Conn) error {
	switch {
	case strings.HasPrefix(command.command, "info"):
		info := rd.info()
		if _, err := conn.Write([]byte(info)); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "ping") || strings.HasPrefix(command.command, "PING"):
		if err := rd.handlePing(command, conn); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "echo"):
		length, args := command.formatArgs()
		if _, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", length, args))); err != nil {
			return err
		}

	case strings.HasPrefix(command.command, "set"):
		if err := rd.setStore(command); err != nil {
			return err
		} else {
			if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
				return err
			}
		}

	case strings.HasPrefix(command.command, "get"):
		if length, value, err := rd.getStore(command); err != nil {
			return err
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
