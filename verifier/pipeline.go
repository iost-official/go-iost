package verifier

type Pipeline struct {
	funcs  []func(interface{}) (interface{}, error)
	middle []chan interface{}
	err    chan error
	stop   bool
}

func (p *Pipeline) Add(f func(interface{}) (interface{}, error), cacheLen int) {
	p.middle = append(p.middle, make(chan interface{}, cacheLen))
	p.funcs = append(p.funcs, f)
}

func (p *Pipeline) End(cl int) {
	p.middle = append(p.middle, make(chan interface{}, cl))
}

func (p *Pipeline) Active() (in, out chan interface{}, err chan error) {
	p.stop = false
	err = make(chan error)
	for i, f := range p.funcs {
		go func() {
			for !p.stop {
				rtn, err0 := f(<-p.middle[i])
				if err != nil {
					err <- err0
				}
				p.middle[i+1] <- rtn
			}

		}()
	}
	return p.middle[0], p.middle[len(p.middle)-1], p.err
}

func (p *Pipeline) Stop() {
	p.stop = true
}
