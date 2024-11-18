package main

import (
	"flag"
	"log"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline"
)

func main() {
	configPath := flag.String("config", "config.yml", "stake pipeline config file")

	flag.Parse()

	conf, err := config.LoadFromFile(*configPath)

	if err != nil {
		log.Fatalf("Unable to load the config: %s", err)
	}

	e, err := pipline.CreateExecutor(conf)

	if err != nil {
		log.Fatalf("Unable to create the pipline executor: %s", err)
	}

	e.Run()
}
