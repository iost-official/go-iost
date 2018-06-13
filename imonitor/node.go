package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"strings"

	"log"

	"github.com/astaxie/beego/config/env"
)

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := redis.Start()
	if err != nil {
		log.Fatal("redis err:\n", err)
	}
	http.HandleFunc("/scripts", handleScripts)
	fmt.Println(http.ListenAndServe("0.0.0.0:30310", nil))
}

func handleScripts(w http.ResponseWriter, r *http.Request) {
	cmd := r.FormValue("cmd")
	daemon := r.FormValue("daemon")
	args := r.FormValue("args")

	script, ok := scripts[cmd]
	if !ok {
		w.Write([]byte("cmd not found"))
		return
	}
	var rsp []byte
	if daemon != "true" {
		rsp = script.run(args)
	} else {
		err := script.daemon(args)
		if err != nil {
			rsp = []byte(err.Error())
		} else {
			rsp = []byte("ok")
		}
	}
	w.Write(rsp)
}

var scripts map[string]script

type script struct {
	run    func(args string) []byte
	daemon func(args string) error
}

var redis, iserver *exec.Cmd

func NewScript(sh string) script {
	return script{
		run: func(args string) []byte {
			var cmd *exec.Cmd
			if len(args) == 0 {
				cmd = exec.Command(sh)
			} else {
				cmd = exec.Command(sh, args)
			}
			rtn, err := cmd.CombinedOutput()
			if err != nil {
				return append([]byte(err.Error()+"\n"), rtn...)
			}
			return rtn
		},
		daemon: func(args string) error {
			var cmd *exec.Cmd
			if len(args) == 0 {
				cmd = exec.Command(sh)
			} else {
				cmd = exec.Command(sh, args)
			}
			err := cmd.Start()
			return err
		},
	}
}

var scriptPath = gopath + "/src/github.com/iost-official/prototype/scripts/"
var gopath = env.Get("GOPATH", "")

func makeScripts() {
	scripts = make(map[string]script)
	scripts["restart-iserver"] = script{
		run: func(args string) []byte {
			iserver := exec.Command(scriptPath+"iserverd.py", "restart")
			err := iserver.Start()
			if err != nil {
				return []byte(err.Error())
			}
			return []byte("ok")
		},
		daemon: func(args string) error {
			iserver := exec.Command(scriptPath+"iserverd.py", "restart")
			err := iserver.Start()
			return err
		},
	}
	scripts["start-iserver"] = script{
		run: func(args string) []byte {
			iserver := exec.Command(scriptPath+"iserverd.py", "start")
			err := iserver.Start()
			if err != nil {
				return []byte(err.Error())
			}
			return []byte("ok")
		},
		daemon: func(args string) error {
			iserver := exec.Command(scriptPath+"iserverd.py", "start")
			err := iserver.Start()
			return err
		},
	}
	scripts["stop-iserver"] = script{
		run: func(args string) []byte {
			iserver := exec.Command(scriptPath+"iserverd.py", "stop")
			err := iserver.Start()
			if err != nil {
				return []byte(err.Error())
			}
			return []byte("ok")
		},
		daemon: func(args string) error {
			iserver := exec.Command(scriptPath+"iserverd.py", "stop")
			err := iserver.Start()
			return err
		},
	}
	scripts["reload"] = script{
		run: func(args string) []byte {
			makeScripts()
			return []byte("ok")
		},
	}

	files, err := ioutil.ReadDir(scriptPath + "entries/")
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			sp := file.Name()
			if !strings.HasPrefix(sp, ".") && strings.HasSuffix(sp, ".sh") {
				scripts[sp] = NewScript(scriptPath + "entries/" + sp)
			}
		}
	}
}

func init() {
	makeScripts()

	redis = exec.Command("redis-server")
	//iserver = exec.Command(scriptPath+"iserverd.py", "start")
}
