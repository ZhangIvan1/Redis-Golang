package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func (rd *Redis) setStore(command Command) error {
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

func (rd *Redis) getStore(command Command) (int, string, error) {
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

func (rd *Redis) handleConnectionTicker() {
	for {
		connection, err := rd.listener.Accept()
		//defer connection.Close()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go rd.handleConnection(connection)
	}
}
