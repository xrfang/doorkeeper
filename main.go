package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dk/base"
	"dk/cli"
	"dk/svr"

	"github.com/mdp/qrterminal"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v2"
)

func main() {
	ver := flag.Bool("version", false, "Show version info")
	cfg := flag.String("conf", "", "Configuration file")
	init := flag.Bool("init", false, "initialize configuration or reset OTP")
	flag.Usage = func() {
		fmt.Printf("DoorKeeper %s\n\n", verinfo())
		fmt.Printf("USAGE: %s [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Printf("OPTIONS:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *ver {
		fmt.Println(verinfo())
		return
	}
	if *init {
		if *cfg == "" {
			fmt.Println("TODO: initialize package like spanx...")
			return
		}
		loadConfig(*cfg)
		if cf.Mode == "server" {
			gopts := totp.GenerateOpts{
				AccountName: cf.Server.OTP.Account,
				Issuer:      cf.Server.OTP.Issuer,
				Algorithm:   otp.AlgorithmSHA256,
			}
			key, err := totp.Generate(gopts)
			assert(err)
			qrterminal.Generate(key.String(), qrterminal.L, os.Stdout)
			cf.Server.OTP.Key = key.Secret()
			f, err := os.Create(*cfg)
			assert(err)
			defer f.Close()
			ye := yaml.NewEncoder(f)
			assert(ye.Encode(&cf))
		} else {
			fmt.Println("OTP key initialization is for DK server only (given client config)")
		}
		return
	}
	if *cfg == "" {
		fmt.Println("ERROR: missing configuration (-conf), try -h for help")
		return
	}
	loadConfig(*cfg)
	base.InitLogger(cf.Logging.Path, cf.Logging.Split, cf.Logging.Keep, cf.Debug)
	if err := ulimit(cf.ULimit); err != nil {
		base.Log("ulimit(): %v", err)
	}
	switch cf.Mode {
	case "client":
		cli.Start(cf.Client)
	case "server":
		if cf.Server.OTP.Key == "" {
			base.Log(`ERROR: missing "server.otp.key", use "-init" to generate`)
			return
		}
		svr.Start(cf.Server)
	}
}
