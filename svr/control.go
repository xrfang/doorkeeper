package svr

import (
	"dk/base"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

/**
访问URL的路径部分即控制命令，格式定义如下：

	/otp/name/port/ip

其中：
  - otp:  一次性密码，由FreeOTP等工具生成
  - name: 需访问的后端的名称
  - port: 目标端口号
  - ip:   目标地址，若不提供，默认为127.0.0.1，即DK客户端所在机器。
		  该值若为*，表示需要查询打开指定端口的内网所有IP。

返回约定：

- 发生在任何错误，例如OTP不正确、后端名字不正确、端口非法等，一律返回HTTP/404。
- 如命令执行成功，返回HTTP/200。内容分两种情况：
  - IP不是"*"：内容分3行，第一行为"OK"；第二行为授权空闲超时时间；第三行为
    授权终止时间。时间格式为RFC3339。
  - IP是"*"：内容首行为"OK"或"ERR"。若为OK，后续行为打开指定端口的IP列表，
	一行一个IP；若为ERR，后续行为错误原因

其他命令：//TODO
- 若仅提供OTP或OTP+name时应该如何？
- 主动取消授权
*/

func controller(cf Config) func(http.ResponseWriter, *http.Request) {
	fwr := regexp.MustCompile(`for=(?:["[])*([0-9a-fA-F:.]+)`)
	return func(w http.ResponseWriter, r *http.Request) {
		s := strings.Split(r.URL.Path[1:], "/")
		switch len(s) {
		case 3:
			s = append(s, "127.0.0.1")
		case 4:
			if s[3] == "" {
				s[3] = "127.0.0.1"
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		port, _ := strconv.Atoi(s[2])
		if port < 0 || port > 65535 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if !totp.Validate(s[0], cf.OTP.Key) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		srch, srcp, _ := net.SplitHostPort(r.RemoteAddr)
		src := srch
		ff := fwr.FindStringSubmatch(r.Header.Get("Forwarded"))
		if len(ff) > 1 {
			src = ff[1]
		}
		srcIP := net.ParseIP(src)
		if srcIP == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if s[3] == "*" {
			b := sm.backends[s[1]]
			if b == nil {
				fmt.Fprintln(w, "ERR")
				fmt.Fprintf(w, "backend \"%s\" not found\n", s[1])
				return
			}
			ip := net.ParseIP(srch)
			sp, _ := strconv.Atoi(srcp)
			qry := base.Chunk{
				Type: base.CT_QRY,
				Src:  &net.TCPAddr{IP: ip, Port: sp},
				Dst:  &net.TCPAddr{IP: srcIP, Port: port},
			}
			err := qry.Send(b.serv)
			if err != nil {
				fmt.Fprintln(w, "ERR")
				fmt.Fprintln(w, err)
				return
			}
			select {
			case msg := <-addChan(srch, sp):
				fmt.Fprint(w, string(msg))
			case <-time.After(time.Minute):
				http.Error(w, "timeout", http.StatusRequestTimeout)
			}
			delChan(srch, sp)
			return
		}
		ip := net.ParseIP(s[3])
		if ip == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		now := time.Now()
		maxIdle := now.Add(time.Duration(cf.IdleClose) * time.Second)
		maxLife := now.Add(time.Duration(cf.AuthTime) * time.Second)
		addr := net.TCPAddr{IP: ip, Port: port}
		ra.Register(src, s[1], &addr)
		fmt.Fprintln(w, "OK")
		fmt.Fprintln(w, maxIdle.Format(time.RFC3339))
		fmt.Fprintln(w, maxLife.Format(time.RFC3339))
	}
}
