package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	DEFAULT_TYPE string = "tcp"
	DEFAULT_HOST string = "0.0.0.0"
	DEFAULT_PORT string = "6379"

	DEFAULT_MASTER_HOST string = ""
	DEFAULT_MASTER_PORT string = ""
)

const (
	MASTER string = "master"
	SLAVE  string = "slave"
)

type Config struct {
	netType string
	host    string
	port    string

	masterHost string
	masterPort string
}

type Request struct {
	Lines    []string
	Commands [][]string
}

type Redis struct {
	listener net.Listener

	store          map[string]string
	timestamp      map[string]time.Time
	timeExpiration map[string]time.Duration

	role string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	args := os.Args

	config := Config{
		netType:    DEFAULT_TYPE,
		host:       DEFAULT_HOST,
		port:       DEFAULT_PORT,
		masterHost: DEFAULT_MASTER_HOST,
		masterPort: DEFAULT_MASTER_PORT,
	}
	for i, arg := range args {
		if arg == "--port" {
			config.port = args[i+1]
		}
		if arg == "--replicaof" {
			config.masterHost = args[i+1]
			config.masterPort = args[i+2]
		}
	}

	fmt.Println("config", config)

	rd := Make(config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("Shutting down the server...")

	if err := rd.listener.Close(); err != nil {
		fmt.Println("Error closing redis server:", err)
	}
}

func (rd *Redis) handleConnection(conn net.Conn) {
	defer conn.Close()

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
				fmt.Println("Now running: " + rd.formatCommand(reqs.Commands[com]))
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
		os.Exit(1)
	}

	req.Lines = strings.Split(string(readBuffer[:n]), "\r\n")

	for line := 0; line < len(req.Lines); line++ {
		fmt.Println(req.Lines[line])
	}

	return req, nil
}

func (rd *Redis) handleResponseLines(reqLine []string, commands *[][]string) error {
	if commands == nil {
		commands = &[][]string{}
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
			*commands = append(*commands, command)
			fmt.Println("inserted command:", rd.formatCommand(command))
			i += 2*n + 1
		default:
			i++
		}
	}

	return nil
}

func (rd *Redis) info() (string, error) {
	//info := "# Replication\r\n"
	role := "role:" + rd.role
	info := "$" + strconv.Itoa(len(role)) + "\r\n" + role + "\r\n"

	return info, nil
}

func (rd *Redis) runCommand(command []string, conn net.Conn) error {
	switch {
	case strings.HasPrefix(command[0], "info"):
		if info, err := rd.info(); err != nil {
			return err
		} else {
			if _, err := conn.Write([]byte(info)); err != nil {
				return err
			}
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

func Make(config Config) *Redis {
	rd := &Redis{}
	rd.store = make(map[string]string)
	rd.timestamp = make(map[string]time.Time)
	rd.timeExpiration = make(map[string]time.Duration)

	if config.masterHost == "" {
		rd.role = MASTER
	} else {
		rd.role = SLAVE
	}

	err := error(nil)
	rd.listener, err = net.Listen(config.netType, config.host+":"+config.port)
	if err != nil {
		log.Fatalln("Failed to bind to port", config.port, err)
	}

	go rd.handleConnectionTicker()

	return rd
}
