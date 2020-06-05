package main

import (
	"fmt"
	"runtime"
	"strings"
)

func assert(e interface{}) {
	switch e.(type) {
	case bool:
		if !e.(bool) {
			panic("assertion failed")
		}
	case nil:
	default:
		panic(e)
	}
}

type exception []string

func (e exception) Error() string {
	return strings.Join(e, "\n")
}

func trace(msg string, args ...interface{}) error {
	pfx := 0
	ex := exception{fmt.Sprintf(msg, args...)}
	n := 1
	for {
		n++
		pc, file, line, ok := runtime.Caller(n)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		name := f.Name()
		if strings.HasPrefix(name, "runtime.") {
			continue
		}
		if pfx == 0 {
			pfx = strings.Index(file, "/PKGNAME/") + 1
		}
		fn := file[pfx:]
		ex = append(ex, fmt.Sprintf("\t(%s:%d) %s", fn, line, name))
	}
	return ex
}
