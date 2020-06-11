package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"dk/base"
	"dk/cli"
	"dk/svr"

	"github.com/mdp/qrterminal"
	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v2"
)

func main() {
	ver := flag.Bool("version", false, "show version info")
	cfg := flag.String("conf", "", "configuration file")
	init := flag.Bool("init", false, "create sample configuration "+
		"(without -conf), or\nreset OTP key (with -conf)")
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
	if *cfg == "" {
		if *init {
			f, err := ioutil.TempFile(".", "dk.yaml.")
			assert(err)
			defer f.Close()
			fmt.Fprintln(f, SAMPLE_CFG)
			fmt.Println("sample configuration:", f.Name())
		} else {
			fmt.Println("ERROR: missing configuration (-conf), try -h for help")
		}
		return
	}
	loadConfig(*cfg)
	if *init {
		if cf.Mode == "server" {
			gopts := totp.GenerateOpts{
				AccountName: cf.Server.OTP.Account,
				Issuer:      cf.Server.OTP.Issuer,
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
