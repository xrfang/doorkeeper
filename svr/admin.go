package svr

import (
	"dk/base"
	"fmt"
	"net/http"
	"os"
	"time"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func startAdminIntf(cf Config) {
	http.HandleFunc("/", controller(cf))
	http.HandleFunc("/help", notFound)
	http.HandleFunc("/help/", helper(cf))
	http.HandleFunc("/auths", notFound)
	http.HandleFunc("/auths/", auths(cf))
	http.HandleFunc("/conns", notFound)
	http.HandleFunc("/conns/", conns(cf))
	adm := http.Server{
		Addr:         fmt.Sprintf(":%d", cf.AdminPort),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		err := adm.ListenAndServe()
		base.Log("startAdminIntf: %v", err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}()
}
