package main

import (
	"flag"
	"fmt"
)

func main() {
	ver := flag.Bool("version", false, "Show version info")
	cfg := flag.String("conf", "", "Configuration file")
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
