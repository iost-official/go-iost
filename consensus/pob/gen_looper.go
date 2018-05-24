package pob

import "time"

type GenLooper interface {
	Run()
	Stop()
}

type genLoopImpl struct {
	net                        *Net
	gen                        Generator
	holder                     *Holder
	view                       WitnessView
	Period, WitnessSize, Order time.Duration
	Exit                       chan bool
}

func (gl *genLoopImpl) waitTime() time.Duration {
	if !gl.view.IsWitness(gl.holder.self) {
		return gl.view.NextDue()
	} else {
		now := time.Duration(time.Now().UnixNano() % (gl.Period * gl.WitnessSize).Nanoseconds())
		if now < gl.Period*gl.Order {
			now += gl.Period * gl.WitnessSize
		}
		return now - gl.Period*gl.Order
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
		}
	}
}

func (gl *genLoopImpl) Stop() {
	gl.Exit <- true
}

func (gl *genLoopImpl) genBlockAndSend() {
	if gl.view.IsPrimary(gl.holder.self) {
		block := gl.gen.NewBlock()
		gl.net.BroadcastBlock(&block)
	}
}
