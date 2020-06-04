package svr

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type backend struct {
	main net.Conn
	clis []net.Conn
	sync.Mutex
}

//TODO: 考虑backend的连接管理

type serviceMgr struct {
	handshake time.Duration
	backends  map[string]*backend
	sync.Mutex
}

func (sm *serviceMgr) Init(cf Config) {
	sm.Lock()
	defer sm.Unlock()
	sm.handshake = time.Duration(cf.Handshake) * time.Second
	sm.backends = make(map[string]*backend)
	//TODO: clean-up routine here
}

func (sm *serviceMgr) Validate(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(sm.handshake))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		err, ok := err.(net.Error)
		if !ok || !err.Timeout() {
			fmt.Printf("TODO: log error %v\n", err)
		} else {
			fmt.Printf("TODO: log timeout waiting handshake\n")
		}
		conn.Close()
		return
	}
	conn.SetReadDeadline(time.Time{})
	fmt.Printf("TODO: validate: received: %x\n", buf[:n])
	return
}

func (sm *serviceMgr) Relay(conn net.Conn, token *accessToken) {
	sm.Lock()
	defer sm.Unlock()
	b := sm.backends[token.dst]
	if b == nil {
		//TODO: log
		conn.Close()
		return
	}
	//TODO: 将连接添加进交换组
}

var sm serviceMgr
