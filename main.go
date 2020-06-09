package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"dk/base"
	"dk/cli"
	"dk/svr"
)

func main() {
	ver := flag.Bool("version", false, "Show version info")
	cfg := flag.String("conf", "", "Configuration file")
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
		svr.Start(cf.Server)
	}
}
