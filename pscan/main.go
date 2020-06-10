package main

func assert(e interface{}) {
	if e != nil {
		panic(e)
	}
}

func main() {
	scan(22, []string{"192.168.90.0/24"})
}
