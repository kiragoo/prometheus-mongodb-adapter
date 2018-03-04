package main

import (
	"os"

	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/sasurai-ossan/prometheus-mongodb-adapter/adapter"
)

var (
	appName    = "prometheus-mongodb-adapter"
	appUsage   = ""
	appVersion = "undefined"
)

var appHelpTemplate = `NAME:
   {{.Name}}{{if .VisibleFlags}}

OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

var (
	urlString  string
	database   string
	collection string
	address    string
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Version = appVersion
	app.CustomAppHelpTemplate = appHelpTemplate
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "mongo-url,m",
			EnvVar:      "MONGO_URL",
			Value:       "mongodb://localhost:27017/prometheus",
			Destination: &urlString,
		},
		cli.StringFlag{
			Name:        "database,d",
			EnvVar:      "DATABASE_NAME",
			Value:       "prometheus",
			Destination: &database,
		},
		cli.StringFlag{
			Name:        "collection,c",
			EnvVar:      "COLLECTION_NAME",
			Value:       "prometheus",
			Destination: &collection,
		},
		cli.StringFlag{
			Name:        "address,a",
			EnvVar:      "LISTEN_ADDRESS",
			Value:       "0.0.0.0:8080",
			Destination: &address,
		},
	}
	app.Action = func(c *cli.Context) error {
		mongo, err := adapter.New(urlString, database, collection)
		if err != nil {
			logrus.Error(err)
			return cli.NewExitError("init error", 2)
		}
		defer mongo.Close()

		if err := mongo.Run(address); err != nil {
			logrus.Error(err)
			return cli.NewExitError("listen error", 3)
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
