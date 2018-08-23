package common

// Service defines APIs of resident goroutines.
type Service interface {
	Start() error
	Stop()
}

// App is a list of Service.
type App []Service

// Start starts all the service. It breaks and returns an error if any error occurs.
func (a *App) Start() error {
	for _, s := range *a {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops all the service.
func (a *App) Stop() {
	for _, s := range *a {
		s.Stop()
	}
}
