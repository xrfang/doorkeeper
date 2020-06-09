package svr

import (
	"dk/base"
	"fmt"
	"net/http"
	"os"
	"time"
)

func startAdminIntf(cf Config) {
	http.HandleFunc("/", controller(cf))
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
