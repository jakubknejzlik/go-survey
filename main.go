package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "d,database",
			EnvVar: "DATABASE_URL",
			Value:  "sqlite3://:memory:",
		},
		cli.IntFlag{
			Name:   "p,port",
			EnvVar: "PORT",
			Value:  80,
		},
	}
	app.Action = func(c *cli.Context) error {
		db, err := NewDB(c.String("database"))
		if err != nil {
			return cli.NewExitError(err, 1)
		}

		r := getRouter(db)
		port := strconv.Itoa(c.Int("port"))
		log.Println("Listening...", port)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			return cli.NewExitError(err, 1)
		}

		return nil
	}

	app.Run(os.Args)
}
