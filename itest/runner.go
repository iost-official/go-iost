package itest

// Runner is the benchmark runner
type Runner struct {
	dfile string
}

// NewRunner will return a new runner
func NewRunner(dfile string) *Runner {
	return nil
}

// Run will run the benchmark
func (r *Runner) Run() error {
	return nil
}

// Done will return nil until the runner is done
func (r *Runner) Done() <-chan struct{} {
	return nil
}

// Err will return the error of runner
func (r *Runner) Err() error {
	return nil
}
