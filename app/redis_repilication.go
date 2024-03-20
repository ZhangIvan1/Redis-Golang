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

type Master struct {
	host    string
	port    string
	toSlave net.Conn
}

func (rd *Redis) doReplication(command Command) error {
	for slave := 0; slave < len(rd.replicationSet); slave++ {
		if _, err := rd.replicationSet[slave].toSlave.Write([]byte(command.buildRequest())); err != nil {
			return err
		}
	}
	return nil
}

func (cm *Command) buildRequest() string {
	cnt := 1 + len(cm.args)
	res := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", cnt, len(cm.command), cm.command)
	for arg := 0; arg < len(cm.args); arg++ {
		res += fmt.Sprintf("$%d\r\n%s\r\n", len(cm.args[arg]), cm.args[arg])
	}

	fmt.Println("build request:", res)

	return res
}
