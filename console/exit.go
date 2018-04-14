package console

func Exit() Cmd {
	e := Cmd{
		Name:  "exit",
		Usage: `Stop daemon and quit`,
	}

	e.Exec = func(host *Console, args []string) string {
		host.Stop()
		return "bye"
	}
	return e
}
