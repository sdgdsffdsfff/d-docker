package util

import (
	"net"
	"strings"
)

const rootServer = "192.41.0.4"

//return local ip address
func GetLocalIp() (string , error) {
	
	conn ,err :=  net.Dial("udp", rootServer+":1")
	if err != nil {
		return "", err
	}
	
	return strings.Split(conn.LocalAddr().String(),":")[0],nil
}