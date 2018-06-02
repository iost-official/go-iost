package pob

import "time"

type GenLooper interface {
	Run()
	Stop()
}

type genLoopImpl struct {
	net  *Net
	gen  Generator
	view WitnessView
	Exit chan bool
}

func (gl *genLoopImpl) waitTime() time.Duration {
	if !gl.view.IsWitness(Data.self) {
		return gl.view.NextDue()
	} else {
		now := time.Duration(time.Now().UnixNano() % (Period * WitnessSize).Nanoseconds())
		if now < Period*Order {
			now += Period * WitnessSize
		}
		return now - Period*Order
	}
}

func (gl *genLoopImpl) Run() {
	gl.genBlockAndSend()
	for {
		select {
		case <-gl.Exit:
			return
		case <-time.After(gl.waitTime()):
			gl.genBlockAndSend()
			gl.view.Next()
		}
	}
}

func (gl *genLoopImpl) Stop() {
	gl.Exit <- true
}

func (gl *genLoopImpl) genBlockAndSend() {
	if gl.view.IsPrimary(Data.self) {
		block := gl.gen.NewBlock()
		gl.net.BroadcastBlock(&block)
	}
}
