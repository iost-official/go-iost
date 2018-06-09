package metrics

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func NewServer(port string) {
	log.Infof("Metrics server start!")

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Errorf("Metrics server failed, stop the server: %v", err)
		}
	}()
}
