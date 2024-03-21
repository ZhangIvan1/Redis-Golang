package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func (rd *Redis) setKey(command Command) error {
	rd.store[command.args[0]] = command.args[1]
	rd.timestamp[command.args[0]] = time.Now()

	if len(command.args) > 2 {
		switch {
		case strings.HasPrefix(command.args[2], "px"):
			if millisecond, err := strconv.Atoi(command.args[3]); err != nil {
				return err
			} else {
				rd.timeExpiration[command.args[0]] = time.Duration(millisecond) * time.Millisecond
			}
		}
	}

	fmt.Printf("set [%s]%s\n", command.args[0], command.args[1])

	return nil
}

func (rd *Redis) getKey(command Command) (int, string, error) {
	if _, exists := rd.store[command.args[0]]; !exists {
		return -2, "", errors.New("no key \"" + command.args[0] + "\" found")
	}

	if expiryTime, exists := rd.timeExpiration[command.args[0]]; exists {
		if time.Since(rd.timestamp[command.args[0]]) > expiryTime {
			delete(rd.store, command.args[0])
			delete(rd.timestamp, command.args[0])
			delete(rd.timeExpiration, command.args[0])
			return -1, "", nil
		}
	}

	return len(rd.store[command.args[0]]), rd.store[command.args[0]], nil
}

func (rd *Redis) writeTicker(writeChan chan Pair[Command, net.Conn]) {
	for {
		for writeOption := range writeChan {
			writeCommand, conn := writeOption()
			if err := rd.setKey(writeCommand); err != nil {
				log.Println(err.Error())
			} else {
				sendItem := func(response string, conn net.Conn) func() (string, net.Conn) {
					return func() (string, net.Conn) {
						return response, conn
					}
				}
				rd.sendChan <- sendItem("+OK\r\n", conn)
			}
		}
	}
}

func (rd *Redis) readTicker(readChan chan Pair[Command, net.Conn]) {
	for {
		for readOption := range readChan {
			readCommand, conn := readOption()
			if length, value, err := rd.getKey(readCommand); err != nil {
				log.Println(err.Error())
			} else {
				sendItem := func(response string, conn net.Conn) func() (string, net.Conn) {
					return func() (string, net.Conn) {
						return response, conn
					}
				}
				if length != -1 {
					rd.sendChan <- sendItem(fmt.Sprintf("$%d\r\n%s\r\n", length, value), conn)
				} else {
					rd.sendChan <- sendItem(fmt.Sprintf("$%d\r\n", length), conn)
				}
			}
		}
	}
}
