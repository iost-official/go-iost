package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/network"
)

type Console struct {
	cmds    []Cmd
	running bool
}

var Wg sync.WaitGroup
var Done = make(chan struct{})
var Nn *network.NaiveNetwork
var Db *db.LDBDatabase

func (c *Console) Init(cmds ...Cmd) error {
	c.cmds = make([]Cmd, 0)
	c.running = true
	for _, cc := range cmds {
		c.RegisterCmd(cc)
	}
	return nil
}

func (c *Console) RegisterCmd(cmd Cmd) {
	c.cmds = append(c.cmds, cmd)
}

func (c *Console) Listen(prompt string) {
	for c.running {
		var cmd string
		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		cmd, _ = reader.ReadString('\n')
		args := strings.Fields(cmd)
		if len(args) == 0 {
			continue
		}
		fmt.Print(c.Run(args[0], args[1:]))
	}
}

func (c *Console) Run(name string, args []string) string {
	name = strings.ToLower(name)

	for _, cmd := range c.cmds {
		if cmd.Name != name {
			continue
		}
		return cmd.Exec(c, args)
	}
	return "Command not found\n"
}

func (c *Console) Stop() {
	c.running = false
}

type Cmd struct {
	Name  string
	Usage string
	Exec  func(host *Console, args []string) string
}
