package svr

import (
	"dk/base"
	"fmt"
	"net"
	"time"
)

type Config struct {
	AdminPort int `yaml:"admin_port"`
	ServePort int `yaml:"serve_port"`
	Handshake int `yaml:"handshake"`
	IdleClose int `yaml:"idle_close"`
	AuthTime  int `yaml:"auth_time"`
	OTP       struct {
		Account string `yaml:"account"`
		Issuer  string `yaml:"issuer"`
		Key     string `yaml:"key"`
	} `yaml:"otp"`
	Auth map[string]string `yaml:"auth"`
}

func Start(cf Config) {
	sm.Init(cf)
	ra.Init(cf)
	startAdminIntf(cf)

	//TODO: 测试：允许访问指定服务
	ip := net.ParseIP("192.168.90.54")
	ra.Register("127.0.0.1", "dev", &net.TCPAddr{IP: ip, Port: 22})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cf.ServePort))
	assert(err)
	for {
		conn, err := ln.Accept()
		if err != nil {
			base.Log("accept: %v", err)
			time.Sleep(time.Second)
			continue
		}
		go func(c net.Conn) {
			if t := ra.Connect(c); t != nil {
				sm.Relay(c, t)
				return
			}
			sm.Validate(c)
		}(conn)
	}
}
