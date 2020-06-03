package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type config struct {
	Mode   string `yaml:"mode"`
	Server struct {
		AdminPort int               `yaml:"admin_port"`
		ServePort int               `yaml:"serve_port"`
		Handshake int               `yaml:"handshake"`
		IdleClose int               `yaml:"idle_close"`
		OTPKey    string            `yaml:"otp_key"`
		Auth      map[string]string `yaml:"auth"`
	} `yaml:""`
	Client struct {
		Name    string `yaml:"name"`
		SvrHost string `yaml:"svr_host"`
		SvrPort int    `yaml:"svr_port"`
		Auth    string `yaml:"auth"`
		MaxConn int    `yaml:"max_conn"`
	} `yaml:""`
}

var cf config

func loadConfig(fn string) {
	f, err := os.Open(fn)
	assert(err)
	defer f.Close()
	assert(yaml.NewDecoder(f).Decode(&cf))
	nr := regexp.MustCompile(`(?i)^[a-z0-9.-]{1,32}$`)
	cf.Mode = strings.ToLower(cf.Mode)
	switch cf.Mode {
	case "client":
		if !nr.MatchString(cf.Client.Name) {
			panic(fmt.Errorf("loadConfig: client.name must be 1~32 chars of alphanum, - or ."))
		}
		cf.Client.Name = strings.ToLower(cf.Client.Name)
		if cf.Client.SvrPort <= 0 || cf.Client.SvrPort > 65535 {
			cf.Client.SvrPort = 35357
		}
		if cf.Client.MaxConn <= 0 || cf.Client.MaxConn > 100 {
			cf.Client.MaxConn = 9
		}
	case "server":
		if cf.Server.AdminPort <= 0 || cf.Server.AdminPort > 65535 {
			cf.Server.AdminPort = 3535
		}
		if cf.Server.ServePort <= 0 || cf.Server.ServePort > 65535 {
			cf.Server.ServePort = 35357
		}
		if cf.Server.Handshake <= 0 || cf.Server.Handshake > 60 {
			cf.Server.Handshake = 10
		}
		if cf.Server.IdleClose <= 0 || cf.Server.IdleClose > 3600 {
			cf.Server.IdleClose = 600
		}
		if cf.Server.OTPKey == "" {
			panic(fmt.Errorf("loadConfig: server.otp_key not set"))
		}
		if cf.Server.Auth == nil || len(cf.Server.Auth) == 0 {
			panic(fmt.Errorf("loadConfig: server.auth not set"))
		}
		auth := make(map[string]string)
		for k, v := range cf.Server.Auth {
			k = strings.TrimSpace(strings.ToLower(k))
			v = strings.TrimSpace(v)
			if !nr.MatchString(k) {
				panic(fmt.Errorf("loadConfig: server.auth invalid key `%s`", k))
			}
			auth[k] = v
		}
	default:
		panic(fmt.Errorf(`loadConfig: mode must be "client" or "server"`))
	}
}
