package main

import (
	"log"
	"os"

	"github.com/frimin/pactus-staker/config"
	"github.com/frimin/pactus-staker/pipline"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "pactus-staker",
		Usage: "manage and run pactus staker pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yml",
				Usage:   "stake pipeline config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run the staker pipeline",
				Action: func(c *cli.Context) error {
					configPath := c.String("config")
					conf, err := config.LoadFromFile(configPath)
					if err != nil {
						log.Fatalf("Unable to load the config: %s", err)
					}

					e, err := pipline.CreateExecutor(conf)
					if err != nil {
						log.Fatalf("Unable to create the pipline executor: %s", err)
					}

					return e.Run()
				},
			},
			{
				Name:  "csv",
				Usage: "export all validator addresses from configured wallets to a csv file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "validators.csv",
						Usage:   "output csv file path",
					},
				},
				Action: func(c *cli.Context) error {
					configPath := c.String("config")
					conf, err := config.LoadFromFile(configPath)
					if err != nil {
						log.Fatalf("Unable to load the config: %s", err)
					}

					e, err := pipline.CreateExecutor(conf)
					if err != nil {
						log.Fatalf("Unable to create the pipline executor: %s", err)
					}

					outputFile := c.String("output")
					err = e.ExportValidatorsCsv(outputFile)
					if err != nil {
						log.Fatalf("Failed to export validators to CSV: %s", err)
					}

					log.Printf("Successfully exported validator addresses to %s", outputFile)
					return nil
				},
			},
		},
		DefaultCommand: "run",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
