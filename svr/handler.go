package svr

import (
	"fmt"
	"net"
)

func handle(conn net.Conn) {
	fmt.Println("REMOTE:", conn.RemoteAddr())
	conn.Close()
	//TODO...
}
