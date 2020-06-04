package cli

import (
	"fmt"
	"net"
	"time"
)

type Config struct {
	Name    string `yaml:"name"`
	SvrHost string `yaml:"svr_host"`
	SvrPort int    `yaml:"svr_port"`
	Auth    string `yaml:"auth"`
	MaxConn int    `yaml:"max_conn"`
}

func connect(addr string) {
	c, err := net.Dial("tcp", addr)
	assert(err)
	fmt.Println("TODO: client connected")
	time.Sleep(2 * time.Second)
	c.Write([]byte("abcde"))
	c.Close()
}

func Start(cf Config) {
	addr := fmt.Sprintf("%s:%d", cf.SvrHost, cf.SvrPort)
	connect(addr)
}
