package main

import "fmt"

func assert(e interface{}) {
	if e != nil {
		panic(e)
	}
}

func main() {
	for _, ip := range scan(22, []string{"192.168.90.0/24"}) {
		fmt.Println(ip)
	}
}
