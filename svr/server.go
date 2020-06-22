package svr

import (
	"dk/base"
	"fmt"
	"net"
	"time"
)

type Config struct {
	AdminPort int  `yaml:"admin_port"`
	ServePort int  `yaml:"serve_port"`
	Handshake int  `yaml:"handshake"`
	IdleClose int  `yaml:"idle_close"`
	AuthTime  int  `yaml:"auth_time"`
	M2MIntf   bool `yaml:"m2m_intf"`
	OTP       struct {
		Issuer string `yaml:"issuer"`
		Key    string `yaml:"key"`
	} `yaml:"otp"`
	Auth map[string]string `yaml:"auth"`
}

func Start(cf Config) {
	sm.Init(cf)
	ra.Init(cf)
	startAdminIntf(cf)
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
