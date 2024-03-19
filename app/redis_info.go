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

func (rd *Redis) info() (string, error) {
	//info := "# Replication\r\n"
	info := ""

	appendInfo(&info, role, rd.role)
	appendInfo(&info, masterReplId, rd.masterReplId)
	appendInfo(&info, masterReplOffset, strconv.Itoa(rd.masterReplOffset))

	fmt.Sprintln("info:", info)
	return info, nil
}

func appendInfo(info *string, key, value string) {
	item := key + ":" + value
	newInfo := *info + "$" + strconv.Itoa(len(item)) + "\r\n" + item + "\r\n"
	*info = newInfo
}
