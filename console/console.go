package console

import (
	"fmt"
	"strings"
)

func Listen() {
	for {
		var cmd string
		fmt.Print("> ")
		fmt.Scanln(&cmd)
		args := strings.Fields(cmd)
		if len(args) == 0 {
			continue
		}
		switch strings.ToLower(args[0]) {
		case "exit":
			fmt.Println("bye!")
			return
		default:
			fmt.Print(Run(args[0], args[1:]))
		}
	}
}

var cmds map[string]Cmd

func RegistCmd(cmd Cmd) {
	cmds[cmd.Name] = cmd
}

func Run(name string, args []string) string {
	name = strings.ToLower(name)
	if cmd, ok := cmds[name]; ok {
		return cmd.Exec(args)
	} else {
		return "Command not found"
	}
}

type Cmd struct {
	Name  string
	Usage string
	Exec  func(args []string) string
}

func init() {
	cmds = make(map[string]Cmd)

	help := Cmd{
		Name:  "help",
		Usage: `Print this manifest`,
		Exec: func(args []string) string {
			s := "======COMMAND LIST\n"
			for k, v := range cmds {
				s += fmt.Sprintf("%v :\n\t%v\n", k, v.Usage)
			}

			s += "exit :\n\tStop daemon and quit\n"
			return s
		},
	}
	RegistCmd(help)
}

