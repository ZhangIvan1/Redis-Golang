package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
)

func (rd *Redis) sendPing(conn net.Conn) {
	if _, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}
	rd.sendReplConf(conn)
}

func (rd *Redis) sendReplConf(conn net.Conn) {
	if _, err := conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n" + rd.config.port + "\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}
	if _, err := conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}

	rd.sendPSync(conn)
}

func (rd *Redis) sendPSync(conn net.Conn) {
	if _, err := conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}
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

		if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
			return err
		}
	} else {
		for i := 0; i < len(command.args); i++ {
			if command.args[i] == "GETACK" {
				if _, err := conn.Write([]byte(fmt.Sprintf(
					"*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%d\r\n",
					len(strconv.Itoa(rd.masterReplOffset)),
					rd.masterReplOffset,
				))); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (rd *Redis) handlePSync(command Command, conn net.Conn) error {
	if _, err := conn.Write([]byte("+FULLRESYNC " + rd.masterReplId + " " + strconv.Itoa(rd.masterReplOffset) + "\r\n")); err != nil {
		return err
	}
	return rd.sendRDB(conn)
}

func (rd *Redis) sendRDB(conn net.Conn) error {
	RDBContents :=
		"524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	if contents, err := hex.DecodeString(RDBContents); err != nil {
		return err
	} else {
		if _, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(contents), contents))); err != nil {
			return err
		}
	}

	return nil
}

func (rd *Redis) handlePing(command Command, conn net.Conn) error {
	if rd.role == MASTER {
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			return err
		}
	}
	return nil
}
