package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func (rd *Redis) sendPing(conn net.Conn) {
	rd.sendChan <- NewPair("*1\r\n$4\r\nping\r\n", conn)

	if err := rd.connectionPool.putConn(conn); err != nil {
		log.Println("Error occur during insert to connectionPool:", err.Error())
		return
	}
	rd.sendReplConf(conn)
}

func (rd *Redis) sendReplConf(conn net.Conn) {
	rd.sendChan <- NewPair(fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%s\r\n", rd.config.port), conn)
	rd.sendChan <- NewPair("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n", conn)

	rd.sendPSync(conn)
}

func (rd *Redis) sendPSync(conn net.Conn) {
	rd.sendChan <- NewPair("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n", conn)
}

func (rd *Redis) handshakeTicker() {
	conn, err := net.Dial(DEFAULT_TYPE, rd.masterHost+":"+rd.masterPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err.Error())
		os.Exit(1)
	}
	rd.masterConn = conn
	go rd.sendPing(conn)
}

func (rd *Redis) handleReplConf(command Command, conn net.Conn) error {
	if rd.role == MASTER {
		for i := 0; i < len(command.args); i++ {
			newSlave := Slave{toSlave: conn}
			if command.args[i] == "listening-port" {
				newSlave.port = command.args[i+1]
				rd.replicationSet = append(rd.replicationSet, newSlave)
			}
		}

		rd.sendChan <- NewPair("+OK\r\n", conn)
	} else {
		for i := 0; i < len(command.args); i++ {
			if command.args[i] == "GETACK" {
				rd.masterReplOffset += command.commandOffset
				rd.sendChan <- NewPair(
					fmt.Sprintf(
						"*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%d\r\n",
						len(strconv.Itoa(rd.masterReplOffset)),
						rd.masterReplOffset,
					),
					conn,
				)
			}
		}
	}

	return nil
}

func (rd *Redis) handlePSync(command Command, conn net.Conn) error {
	rd.sendChan <- NewPair("+FULLRESYNC "+rd.masterReplId+" "+strconv.Itoa(rd.masterReplOffset)+"\r\n", conn)
	return rd.sendRDB(conn)
}

func (rd *Redis) sendRDB(conn net.Conn) error {
	RDBContents :=
		"524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	if contents, err := hex.DecodeString(RDBContents); err != nil {
		return err
	} else {
		rd.sendChan <- NewPair(fmt.Sprintf("$%d\r\n%s", len(contents), contents), conn)
	}

	return nil
}

func (rd *Redis) handlePing(command Command, conn net.Conn) error {
	if rd.role == MASTER {
		rd.sendChan <- NewPair("+PONG\r\n", conn)
	} else if rd.role == SLAVE {
		rd.masterReplOffset += command.commandOffset
		if rd.masterConn != conn {
			rd.masterConn.Close()
			rd.connectionPool.removeConn(rd.masterConn)
			rd.masterConn = conn
			rd.connectionPool.putConn(rd.masterConn)
		}
	}
	return nil
}
