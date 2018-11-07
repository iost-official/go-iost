package itest

type Runner struct {
	dfile string
}

func NewRunner(dfile string) *Runner {
	return nil
}

func (r *Runner) Run() error {
	return nil
}

func (r *Runner) Done() <-chan struct{} {
	return nil
}

func (r *Runner) Err() error {
	return nil
}
