package main

import (
	"log"
	"os"

	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
)

func main() {
	cfg, err := cli.LoadConfig(os.Args[1:])

	if err != nil {
		log.Fatal("config error: ", err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}
