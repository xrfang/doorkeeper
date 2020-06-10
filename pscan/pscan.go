package main

import (
	"net"
	"sync"
)

func scan(port int, cidrs []string) (ips []string) {
	resc := make(chan string)
	go func() { //收集结果
		for {
			ip := <-resc
			if ip == "" {
				break
			}
			ips = append(ips, ip)
		}
	}()
	const THREADS = 100 //为简单起见，固定开100个线程
	task := make(chan string, THREADS)
	var wg sync.WaitGroup
	for i := 0; i < THREADS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				ip := <-task
				if ip == "" {
					break
				}
				//TODO：do the scan...
				func ScanPort(ip string, port int, timeout time.Duration) {
					target := fmt.Sprintf("%s:%d", ip, port)
					conn, err := net.DialTimeout("tcp", target, timeout)
					if err != nil {
						if strings.Contains(err.Error(), "too many open files") {
							time.Sleep(timeout)
							ScanPort(ip, port, timeout)
						} else {
							fmt.Println(port, "closed")
						}
						return
					}
				
					conn.Close()
					fmt.Println(port, "open")
				}
				resc <- ip
			}
		}()
	}
	inc := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}
	for _, cidr := range cidrs {
		ip, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
			task <- ip.String()
		}
	}
	for i := 0; i < THREADS; i++ {
		task <- ""
	}
	wg.Wait()  //等待所有工作线程结束
	resc <- "" //通知收集线程结束
	return
}
