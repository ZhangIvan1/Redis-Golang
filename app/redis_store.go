package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func (rd *Redis) setStore(command []string) error {
	rd.store[command[1]] = command[2]
	rd.timestamp[command[1]] = time.Now()

	if len(command) > 3 {
		switch {
		case strings.HasPrefix(command[3], "px"):
			if millisecond, err := strconv.Atoi(command[4]); err != nil {
				return err
			} else {
				rd.timeExpiration[command[1]] = time.Duration(millisecond) * time.Millisecond
			}
		}
	}

	return nil
}

func (rd *Redis) getStore(key string) (string, string, error) {
	if _, exists := rd.store[key]; !exists {
		return "", "", errors.New("no key \"" + key + "\" found")
	}

	if expiryTime, exists := rd.timeExpiration[key]; exists {
		if time.Since(rd.timestamp[key]) > expiryTime {
			delete(rd.store, key)
			delete(rd.timestamp, key)
			delete(rd.timeExpiration, key)
			return "$-1\r\n", "", nil
		}
	}

	return "$" + strconv.Itoa(len(rd.store[key])) + "\r\n", rd.store[key] + "\r\n", nil
}

func (rd *Redis) handleConnectionTicker() {
	for {
		connection, err := rd.listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go rd.handleConnection(connection)
	}
}
