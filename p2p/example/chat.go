package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
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
	From    string
	ID      string
	Content string
}

func newMessage(content string, from string) *message {
	id, _ := uuid.NewV4()
	return &message{
		ID:      id.String(),
		Content: content,
		From:    from,
	}
}

// Chatter is the core struct of the chatting demo.
type Chatter struct {
	p2pService *p2p.NetService

	msg    chan p2p.IncomingMessage
	quitCh chan struct{}

	mu       sync.RWMutex
	received map[string]struct{}
}

// NewChatter returns a new instance of Chatter.
func NewChatter(p2pService *p2p.NetService) *Chatter {
	c := &Chatter{
		p2pService: p2pService,
		quitCh:     make(chan struct{}),
		received:   make(map[string]struct{}),
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
	ilog.Infof("chatter is stopped")
	close(ct.quitCh)
}

func (ct *Chatter) handleMsgLoop() {
	for {
		select {
		case <-ct.quitCh:
			ilog.Infof("chatter quits loop")
			return
		case msg := <-ct.msg:
			var m message
			err := json.Unmarshal(msg.Data(), &m)
			if err != nil {
				ilog.Errorf("json decode failed. err=%v, bytes=%v", err, msg.Data())
				continue
			}
			if ct.hasID(m.ID) {
				continue
			}
			author := shortID(m.From) + ":"
			ct.clearLastLine(0)
			fmt.Println(color(author, green), color(m.Content, blue))
			// ct.reprint()
			ct.printPrompt()
			ct.record(m.ID)
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
			ilog.Errorf("std read error. err=%v", err)
			panic(err)
		}
		if sendData[len(sendData)-1] == '\n' {
			sendData = sendData[0 : len(sendData)-1]
		}
		if len(sendData) == 0 {
			continue
		}
		msg := newMessage(sendData, ct.p2pService.ID())
		bytes, err := json.Marshal(msg)
		if err != nil {
			ilog.Errorf("json encode failed. err=%v, obj=%+v", err, msg)
			continue
		}
		ct.record(msg.ID)
		ct.p2pService.Broadcast(bytes, chatData, p2p.UrgentMessage)
		author := shortID(ct.p2pService.ID()) + ":"
		ct.clearLastLine(1)
		fmt.Println(color(author, yellow), color(sendData, blue))
	}
}

func (ct *Chatter) reprint() {
	// robotgo.KeyTap("r", "control")
}

func (ct *Chatter) clearLastLine(n int) {
	if n > 0 {
		fmt.Printf("\033[%dA", n)
	}
	fmt.Print("\033[2K")
	fmt.Print("\033[1G")
}

func (ct *Chatter) printPrompt() {
	fmt.Print("\033[05;0m> \033[0m")
}

func (ct *Chatter) printOpening() {
	text := fmt.Sprintf(opening, shortID(ct.p2pService.ID()), ct.p2pService.LocalAddrs()[0], ct.p2pService.ID())
	fmt.Println(color(text, grey))
}

func (ct *Chatter) record(id string) {
	ct.mu.Lock()
	ct.received[id] = struct{}{}
	if len(ct.received) > 100000 {
		ct.received = make(map[string]struct{})
	}
	ct.mu.Unlock()
}

func (ct *Chatter) hasID(id string) bool {
	ct.mu.RLock()
	_, exist := ct.received[id]
	ct.mu.RUnlock()
	return exist
}
