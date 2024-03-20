package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
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
	for {
		conn, err := net.Dial(DEFAULT_TYPE, rd.masterHost+":"+rd.masterPort)
		if err != nil {
			fmt.Println("Error connecting to master:", err.Error())
			os.Exit(1)
		}

		go rd.sendPing(conn)

		go time.Sleep(20 * time.Millisecond)
	}
}

func (rd *Redis) handleReplConf(conn net.Conn, command []string) error {
	if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
		return err
	}
	return nil
}

func (rd *Redis) handlePSync(conn net.Conn, command []string) error {
	if _, err := conn.Write([]byte("+FULLRESYNC " + rd.masterReplId + " " + strconv.Itoa(rd.masterReplOffset) + "\r\n")); err != nil {
		return err
	}
	return rd.sendRDB(conn)
}

func (rd *Redis) sendRDB(conn net.Conn) error {
	contents := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	if _, err := conn.Write([]byte("$" + strconv.Itoa(len(contents)) + "\r\n" + contents)); err != nil {
		return err
	}
	return nil
}
