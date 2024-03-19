package main

import (
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

	return "$" + strconv.Itoa(info.length-2) + "\r\n" + info.info
}

func (info *Info) appendInfo(key, value string) {
	item := key + ":" + value
	info.length += len(item)
	info.info += item + "\r\n"
}
