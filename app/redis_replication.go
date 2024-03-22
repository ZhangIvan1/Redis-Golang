package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type Slave struct {
	host    string
	port    string
	toSlave net.Conn
}

func (rd *Redis) doReplication(command Command) {
	fmt.Printf("i have %d to send\n", len(rd.replicationSet))
	for _, slave := range rd.replicationSet {
		fmt.Printf("sent replication to: %s\n", slave.port)
		rd.sendChan <- NewPair(command.buildRequest(), slave.toSlave)
	}
}

func (cm *Command) buildRequest() string {
	cnt := 1 + len(cm.args)
	res := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", cnt, len(cm.command), cm.command)
	for arg := 0; arg < len(cm.args); arg++ {
		res += fmt.Sprintf("$%d\r\n%s\r\n", len(cm.args[arg]), cm.args[arg])
	}

	return res
}

func (rd *Redis) handleWait(command Command, conn net.Conn) {
	go func() {
		if rd.role == MASTER {
			callTime := time.Now()
			numreplicas, _ := strconv.Atoi(command.args[0])
			waitTimeout, _ := strconv.Atoi(command.args[1])
			for {
				if time.Since(callTime) >= time.Duration(waitTimeout) || numreplicas <= len(rd.replicationSet) {
					rd.sendChan <- NewPair(fmt.Sprintf(":%d\r\n", len(rd.replicationSet)), conn)
				}
			}
		}
	}()
}
