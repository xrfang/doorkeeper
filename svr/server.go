package svr

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type Config struct {
	AdminPort int               `yaml:"admin_port"`
	ServePort int               `yaml:"serve_port"`
	Handshake int               `yaml:"handshake"`
	IdleClose int               `yaml:"idle_close"`
	AuthTime  int               `yaml:"auth_time"`
	OTPKey    string            `yaml:"otp_key"`
	Auth      map[string]string `yaml:"auth"`
	TLSCert   string            `yaml:"tls_cert"`
	TLSPKey   string            `yaml:"tls_pkey"`
}

func Start(cf Config) {
	adm := http.Server{
		Addr:         fmt.Sprintf(":%d", cf.AdminPort),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		if cf.TLSCert != "" && cf.TLSPKey != "" {
			assert(adm.ListenAndServeTLS(cf.TLSCert, cf.TLSPKey))
		} else {
			assert(adm.ListenAndServe())
		}
	}()
	sm.Init(cf)
	ra.Init(cf)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cf.ServePort))
	assert(err)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("TODO: log error:", err)
			time.Sleep(time.Second)
			continue
		}
		go func(c net.Conn) {
			if ra.Connect(c) {
				//TODO: RELAY...
			}
			sm.Validate(c)
		}(conn)
	}
}
