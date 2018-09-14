package metrics

var defaultClient *Client

func init() {
	defaultClient = NewClient()
}

// InitMetrics inits the defaultClient.
func InitMetrics(c *Client) {
	defaultClient.Stop()
	defaultClient = c
}

// SetPusher sets the pusher with the given addr.
func SetPusher(addr, username, password string) error {
	return defaultClient.SetPusher(addr, username, password)
}

// SetID sets the ID of metrics client.
func SetID(id string) {
	defaultClient.SetID(id)
}

// Start starts the pusher loop.
func Start() error {
	return defaultClient.Start()
}

// Stop stops the pusher loop.
func Stop() {
	defaultClient.Stop()
}

// NewCounter returns a counter-type metrics.
func NewCounter(name string, labels []string) Counter {
	return defaultClient.NewCounter(name, labels)
}

// NewGauge returns a gauge-type metrics.
func NewGauge(name string, labels []string) Gauge {
	return defaultClient.NewGauge(name, labels)
}

// NewSummary returns a summary-type metrics.
func NewSummary(name string, labels []string) Summary {
	return defaultClient.NewSummary(name, labels)
}
