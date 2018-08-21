package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	uuid "github.com/satori/go.uuid"
)

const (
	chatData p2p.MessageType = 100

	opening = `
	Welcome! Your ID is %s.
	You can serve as a seed node by running "./example --seed %s/ipfs/%s" on another console. You can replace 127.0.0.1 with public IP as well.
	`
)

type message struct {
	ID      string
	Content string
}

func newMessage(content string) *message {
	id, _ := uuid.NewV4()
	return &message{
		ID:      string(id.Bytes()),
		Content: content,
	}
}

// Chatter is the core struct of the chatting demo.
type Chatter struct {
	p2pService *p2p.NetService

	msg        chan p2p.IncomingMessage
	quitCh     chan struct{}
	screenLock sync.Mutex
}

// NewChatter returns a new instance of Chatter.
func NewChatter(p2pService *p2p.NetService) *Chatter {
	c := &Chatter{
		p2pService: p2pService,
		quitCh:     make(chan struct{}),
	}
	c.msg = p2pService.Register("example_chat", chatData)
	return c
}

// Start starts chatter's job.
func (ct *Chatter) Start() {
	ct.printOpening()
	go ct.handleMsgLoop()
	go ct.readLoop()
}

// Stop stops chatter's job.
func (ct *Chatter) Stop() {
	ilog.Info("chatter is stopped")
	close(ct.quitCh)
}

func (ct *Chatter) handleMsgLoop() {
	for {
		select {
		case <-ct.quitCh:
			ilog.Info("chatter quits loop")
			return
		case msg := <-ct.msg:
			var m message
			err := json.Unmarshal(msg.Data(), &m)
			if err != nil {
				ilog.Error("json decode failed. err=%v, bytes=%v", err, msg.Data())
				continue
			}
			author := "\n" + shortID(msg.From().Pretty()) + ": "
			fmt.Println(color(author, green), color(m.Content, blue))
			// fmt.Print("\033[05;0m> \033[0m")
			ct.p2pService.Broadcast(msg.Data(), chatData, p2p.UrgentMessage)
		}
	}
}

func (ct *Chatter) readLoop() {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		ct.printPrompt()
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			ilog.Error("std read error. err=%v", err)
			panic(err)
		}
		if sendData[len(sendData)-1] == '\n' {
			sendData = sendData[0 : len(sendData)-1]
		}
		msg := newMessage(sendData)
		bytes, err := json.Marshal(msg)
		if err != nil {
			ilog.Error("json encode failed. err=%v, obj=%+v", err, msg)
			continue
		}
		ct.p2pService.Broadcast(bytes, chatData, p2p.UrgentMessage)
		author := shortID(ct.p2pService.ID()) + ": "
		fmt.Print(color(author, green), color(sendData, blue))
	}
}

func (ct *Chatter) clearLastLine() {
	fmt.Print("\033[1A")
	fmt.Print("\033[2K")
}

func (ct *Chatter) printPrompt() {
	fmt.Print("\033[05;0m> \033[0m")
}

func (ct *Chatter) printOpening() {
	text := fmt.Sprintf(opening, shortID(ct.p2pService.ID()), ct.p2pService.LocalAddrs()[0], ct.p2pService.ID())
	fmt.Println(color(text, grey))
}
