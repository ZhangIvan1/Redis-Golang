package main

import "strconv"

func (rd *Redis) info() (string, error) {
	//info := "# Replication\r\n"
	role := "role:" + rd.role
	info := "$" + strconv.Itoa(len(role)) + "\r\n" + role + "\r\n"

	return info, nil
}
