package console

func Exit() Cmd {
	e := Cmd{
		Name:  "exit",
		Usage: `Stop daemon and quit`,
	}

	e.Exec = func(host *Console, args []string) string {
		close(Done)
		Wg.Wait()
		host.Stop()
		return "bye"
	}
	return e
}
