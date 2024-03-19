package main

import (
	"fmt"
	"strconv"
)

const (
	role             string = "role"
	masterReplId     string = "master_replid"
	masterReplOffset string = "master_repl_offset"
)

type Info struct {
	length int
	info   string
}

func (rd *Redis) info() string {
	//info := "# Replication\r\n"
	info := &Info{0, ""}

	info.appendInfo(role, rd.role)
	info.appendInfo(masterReplId, rd.masterReplId)
	info.appendInfo(masterReplOffset, strconv.Itoa(rd.masterReplOffset))

	fmt.Sprintln("info:", info)
	return "$" + strconv.Itoa(info.length) + "\r\n" + info.info
}

func (info *Info) appendInfo(key, value string) {
	item := key + ":" + value
	info.length += len(item)
	info.info += item + "\r\n"
}
