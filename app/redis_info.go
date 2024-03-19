package main

import "strconv"

const (
	role               string = "role"
	master_replid      string = "masterReplId"
	master_repl_offset string = "masterReplOffset"
)

func (rd *Redis) info() (string, error) {
	//info := "# Replication\r\n"
	info := ""

	appendInfo(&info, role, rd.role)
	appendInfo(&info, master_replid, rd.masterReplId)
	appendInfo(&info, master_repl_offset, rd.masterReplOffset)

	return info, nil
}

func appendInfo(info *string, key, value string) {
	item := key + ":" + value
	newInfo := *info + "$" + strconv.Itoa(len(item)) + "\r\n" + item + "\r\n"
	*info = newInfo
}
