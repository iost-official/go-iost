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
func SetPusher(addr string) error {
	return defaultClient.SetPusher(addr)
}

// Start starts the pusher loop.
func Start() error {
	return defaultClient.Start()
}

// Stop stops the pusher loop.
func Stop() {
	defaultClient.Stop()
}

// Counter sends a counter type metrics.
func Counter(name string, value float64, tagkv map[string]string) error {
	return defaultClient.Counter(name, value, tagkv)
}

// Gauge sends a gauge type metrics.
func Gauge(name string, value float64, tagkv map[string]string) error {
	return defaultClient.Gauge(name, value, tagkv)
}

// Timer sends a summary type metrics.
func Timer(name string, value float64, tagkv map[string]string) error {
	return defaultClient.Timer(name, value, tagkv)
}
