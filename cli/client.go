package cli

import (
	"dk/utils"
	"fmt"
	"net"
)

type Config struct {
	Name    string `yaml:"name"`
	SvrHost string `yaml:"svr_host"`
	SvrPort int    `yaml:"svr_port"`
	Auth    string `yaml:"auth"`
	MaxConn int    `yaml:"max_conn"`
}

func connect(cf Config) {
	addr := fmt.Sprintf("%s:%d", cf.SvrHost, cf.SvrPort)
	c, err := net.Dial("tcp", addr)
	assert(err)
	fmt.Println("TODO: client connected")
	handshake := utils.Authenticate(nil, cf.Name, cf.Auth)
	c.Write(handshake)
	c.Close()
}

func Start(cf Config) {
	connect(cf)
}
