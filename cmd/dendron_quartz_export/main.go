package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/zuzuka28/dendron-quartz-export/pkg/exporter"

	"gopkg.in/yaml.v2"
)

func main() {
	cfgPath := flag.String("cfg", "", "config path")

	flag.Parse()

	if *cfgPath == "" {
		log.Fatal("config path required")
	}

	raw, err := os.ReadFile(*cfgPath)
	if err != nil {
		log.Fatalf("read config: %s", err.Error())
	}

	cfg := new(exporter.Config)
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		log.Fatalf("Unmarshal config: %s", err.Error())
	}

	e := exporter.New(cfg)

	if err := e.Run(context.Background()); err != nil {
		log.Fatalf("export notes: %s", err.Error())
	}

	log.Println("export done")
}
