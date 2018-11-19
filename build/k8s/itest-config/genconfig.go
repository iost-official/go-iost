package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/iost-official/go-iost/itest"
)

var (
	cluster = flag.String("c", "default", "The name of cluster")
	slaves  = flag.String("s", "7,8", "The number list of slave node")
)

func main() {
	flag.Parse()

	bank := itest.NewAccount(
		"admin",
		"2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1",
		"ed25519",
	)

	clients := make([]*itest.Client, 0)
	for _, slave := range strings.Split(*slaves, ",") {
		client := &itest.Client{
			Name: fmt.Sprintf("iserver-%v", slave),
			Addr: fmt.Sprintf("iserver-%v.iserver.%v.svc.cluster.local:30002", slave, *cluster),
		}
		clients = append(clients, client)
	}

	config := &itest.Config{
		Bank:    bank,
		Clients: clients,
	}

	file, err := os.Create("itest.json")
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Unable to marshal config to JSON: %v", err)
	}
	file.Write(b)
	file.Close()
}
