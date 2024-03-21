package main

import (
	"fmt"
	"net"
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
