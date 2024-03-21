package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConnectionPool struct {
	mu       sync.Mutex
	conns    []net.Conn
	capacity int
}

func (p *ConnectionPool) getConnLock() net.Conn {
	conn := p.conns[0]
	p.conns = p.conns[1:]
	return conn
}

func (p *ConnectionPool) putConn(conn net.Conn) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns) >= p.capacity {
		return fmt.Errorf("connection pool is full")
	}

	p.conns = append(p.conns, conn)
	return nil
}

func (p *ConnectionPool) removeConn(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var newConns []net.Conn
	for _, c := range p.conns {
		if c != conn {
			newConns = append(newConns, c)
		}
	}
	p.conns = newConns
}

type Request struct {
	Lines    []string
	Commands []Command
}

func (rd *Redis) handleConnectionTicker(commandChan chan Pair[Command, net.Conn]) {
	for {
		conn := net.Conn(nil)

		rd.connectionPool.mu.Lock()
		if len(rd.connectionPool.conns) > 0 {
			conn = rd.connectionPool.getConnLock()
		}
		rd.connectionPool.mu.Unlock()

		if conn == nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}

		log.Println("now handling conn:", conn.RemoteAddr())

		data, err := rd.readData(conn)
		if err != nil || data == nil {
			log.Println(err.Error())
			conn.Close()
			rd.connectionPool.removeConn(conn)
			continue
		} else {
			go func() {
				reqs, err := rd.buildRequest(data)
				if err != nil {
					log.Println(err.Error())
					rd.connectionPool.putConn(conn)
					return
				}
				if err := rd.handleResponseLines(reqs.Lines, &reqs.Commands); err != nil {
					log.Println("Error handleResponseLines: ", err.Error())
				}
				for _, command := range reqs.Commands {
					command.commandOffset = len(command.buildRequest())
					commandChan <- NewPair(command, conn)
				}
			}()
		}
	}
}

func (rd *Redis) readData(conn net.Conn) ([]byte, error) {
	readBuffer := make([]byte, 4096)

	n, err := conn.Read(readBuffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Connection", conn.RemoteAddr(), "closed.")
			return nil, err
		} else {
			log.Println("Error reading data from connection", conn.RemoteAddr(), ":", err)
		}
		return nil, err
	}

	if n == 0 {
		log.Println("No data read from connection", conn.RemoteAddr())
		return nil, nil
	}

	return readBuffer[:n], nil
}

func (rd *Redis) buildRequest(data []byte) (req Request, err error) {
	if len(data) == 0 {
		log.Println("No data received to build request")
		return req, nil
	}

	req.Lines = strings.Split(string(data), "\r\n")
	log.Println(req.Lines)

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

func (rd *Redis) listenConnectionTicker() {
	for {
		if connection, err := rd.listener.Accept(); err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		} else {
			log.Println("New connection from:", connection.RemoteAddr().String())
			if err := rd.connectionPool.putConn(connection); err != nil {
				fmt.Println("Error accepting connection:", err.Error())
				return
			}
		}
	}
}

func (rd *Redis) sendTicker(sendChan chan Pair[string, net.Conn]) {
	for {
		for sendOption := range sendChan {
			payload, conn := sendOption()
			if _, err := conn.Write([]byte(payload)); err != nil {
				log.Println(err.Error())
			}
		}
	}
}
