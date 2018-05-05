package console

import "fmt"

func Help() Cmd {
	return Cmd{
		Name:  "help",
		Usage: `Print this manifest`,
		Exec: func(host *Console, args []string) string {
			if host == nil {
				fmt.Println("failed init!!")
			}

			s := "======COMMAND LIST\n"
			for _, v := range host.cmds {
				s += fmt.Sprintf("%v :\n\t%v\n", v.Name, v.Usage)
			}

			return s
		},
	}
}
