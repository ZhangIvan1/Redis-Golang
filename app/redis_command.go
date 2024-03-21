package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
)

type Command struct {
	commandType   string
	commandLength int
	commandOffset int
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

func (cm *Command) buildCommand() {

}

func (rd *Redis) commandClassifyTicker(commandChan chan Pair[Command, net.Conn]) {
	for {
		for commandItem := range commandChan {
			command, conn := commandItem()
			log.Println("get command:", command.formatCommand())

			switch {
			case command.commandType == "+":
				if err := rd.handleRepose(command, conn); err != nil {
					log.Println(err.Error())
				}

			case command.command == "info" || command.command == "INFO":
				rd.sendChan <- NewPair(rd.info(), conn)

			case command.command == "ping" || command.command == "PING":
				if err := rd.handlePing(command, conn); err != nil {
					log.Println(err.Error())
				}

			case command.command == "echo" || command.command == "ECHO":
				args := command.formatArgs()
				rd.sendChan <- NewPair(fmt.Sprintf("$%d\r\n%s\r\n", len(args), args), conn)

			case command.command == "set" || command.command == "SET":
				rd.writeChan <- NewPair(command, conn)
				if rd.role == MASTER {
					rd.doReplication(command)
				}

			case command.command == "get" || command.command == "GET":
				rd.readChan <- NewPair(command, conn)

			case command.command == "REPLCONF" || command.command == "replconf":
				if err := rd.handleReplConf(command, conn); err != nil {
					log.Println(err.Error())
				}

			case command.command == "PSYNC":
				if err := rd.handlePSync(command, conn); err != nil {
					log.Println(err.Error())
				}

			default:
				log.Printf("no matching command: %s", command.formatCommand())
			}

			rd.connectionPool.mu.Lock()
			rd.connectionPool.conns = append(rd.connectionPool.conns, conn)
			rd.connectionPool.mu.Unlock()
		}
	}
}

func (rd *Redis) handleRepose(command Command, conn net.Conn) error {
	switch {
	case command.command == "PONG":
		fmt.Println("Get +PONG")
	case command.command == "OK":
		fmt.Println("Get +OK")
	case command.command == "FULLRESYNC":
		fmt.Println("Get +FULLRESYNC")
		rd.masterReplId = command.args[0]
		if offset, err := strconv.Atoi(command.args[1]); err != nil {
			return err
		} else {
			rd.masterReplOffset = offset
		}
	default:
		return errors.New(fmt.Sprintf("no matching respose: %s", command.formatCommand()))
	}
	return nil
}
