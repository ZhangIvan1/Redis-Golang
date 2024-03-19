package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func (rd *Redis) sendPing(conn net.Conn) {
	if _, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}
}

func (rd *Redis) sendReplConf(conn net.Conn) {
	if _, err := conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n" + rd.masterPort + "\r\n")); err != nil {
		fmt.Println("Error occur during handshaking to master:", err.Error())
		return
	}
	if _, err := conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")); err != nil {
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

		go rd.sendReplConf(conn)

		go time.Sleep(20 * time.Millisecond)
	}
}
