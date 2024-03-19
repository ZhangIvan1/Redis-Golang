package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func (rd *Redis) sendHandshake(conn net.Conn) {
	if _, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n")); err != nil {
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

		go rd.sendHandshake(conn)

		time.Sleep(20 * time.Millisecond)
	}
}
