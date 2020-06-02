package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	ver := flag.Bool("version", false, "Show version info")
	cfg := flag.String("conf", "", "Configuration file")
	flag.Usage = func() {
		fmt.Printf("DoorKeeper %s\n\n", verinfo())
		fmt.Printf("USAGE: %s [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Println("OPTIONS:\n")
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
	fmt.Println("TODO...")
}
