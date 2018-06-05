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
	go func() {
		for {
			iserver = exec.Command(gopath + "/src/github.com/iost-official/prototype/iserver/iserver")
			b, err := iserver.CombinedOutput()
			fmt.Println(string(b), err.Error())
		}
	}()

	http.HandleFunc("/scripts", handleScripts)
	fmt.Println(http.ListenAndServe("127.0.0.1:30310", nil))
}

func handleScripts(w http.ResponseWriter, r *http.Request) {
	cmd := r.FormValue("cmd")
	script, ok := scripts[cmd]
	if !ok {
		w.Write([]byte("cmd not found"))
	}
	rsp := script.run()
	fmt.Println(string(rsp))

	w.Write(rsp)
}

var scripts map[string]script

type script struct {
	run func() []byte
}

var redis, iserver *exec.Cmd

func NewScript(sh string) script {
	return script{
		run: func() []byte {
			fmt.Println("running", sh)
			cmd := exec.Command(sh)
			rtn, err := cmd.CombinedOutput()
			if err != nil {
				return []byte(err.Error())
			}
			return rtn
		},
	}
}

var scriptPath = gopath + "/src/github.com/iost-official/prototype/"
var gopath = env.Get("GOPATH", "")

func makeScripts() {
	scripts = make(map[string]script)
	scripts["restart-iserver"] = script{
		run: func() []byte {
			err := iserver.Process.Kill()
			if err != nil {
				return []byte(err.Error())
			}
			return []byte("ok")
		},
	}
	scripts["reload"] = script{
		run: func() []byte {
			makeScripts()
			return []byte("ok")
		},
	}

	files, err := ioutil.ReadDir(scriptPath)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			sp := file.Name()
			if !strings.HasPrefix(sp, ".") && strings.HasSuffix(sp, ".sh") {
				scripts[sp] = NewScript(scriptPath + sp)
			}
		}
	}
}

func init() {
	makeScripts()

	redis = exec.Command("redis-server")
	iserver = exec.Command(gopath + "/src/github.com/iost-official/prototype/iserver/iserver")
}
