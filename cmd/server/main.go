package main

import (
	"log"
	"os"

	"github.com/a-aleesshin/metrics/internal/platform/buildinfo"
	"github.com/a-aleesshin/metrics/internal/server/transport/cli"
)

// Информация о сборке. Значения по умолчанию перезаписываются на этапе
// компиляции через ldflags -X (main.buildVersion, main.buildDate, main.buildCommit).
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildinfo.Print(buildVersion, buildDate, buildCommit)

	cfg, err := cli.LoadConfig(os.Args[1:])

	if err != nil {
		log.Fatal("config error: ", err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}
