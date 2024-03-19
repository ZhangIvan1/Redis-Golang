package main

import (
	"strconv"
)

const (
	role             string = "role"
	masterReplId     string = "master_replid"
	masterReplOffset string = "master_repl_offset"
)

func (rd *Redis) info() string {
	//info := "# Replication\r\n"
	info := ""

	info = appendInfo(info, role, rd.role)
	info = appendInfo(info, "master_replid", rd.masterReplId)
	info = appendInfo(info, "master_repl_offset", strconv.Itoa(rd.masterReplOffset))

	return info
}

func appendInfo(info, key, value string) string {
	item := key + ":" + value
	return info + "$" + strconv.Itoa(len(item)) + "\r\n" + item + "\r\n"
}
