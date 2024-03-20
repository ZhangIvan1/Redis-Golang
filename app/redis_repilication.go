package main

import "net"

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
