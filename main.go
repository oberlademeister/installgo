package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const DLPageInfoURL = "https://go.dev/dl/?mode=json"
const DLURLBase = "https://go.dev/dl/"

func main() {
	app := &cli.App{
		Name:   "goinstall",
		Usage:  "install golang form main repo into ./go",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dstdir",
				Aliases: []string{"d"},
				Usage:   "valid path where goinstall will unpack tha tar",
				EnvVars: []string{"GOINSTALDSTPATH"},
				Value:   "./go",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.Infof("retrieving download information")
	info, err := RetrievePageResult(DLPageInfoURL)
	if err != nil {
		return err
	}
	latestStable, err := GetLatestStableVersion(info)
	if err != nil {
		return err
	}
	logrus.WithField("latestStableVersion", latestStable.String()).Infof("information retrieved")
	fileInfo := SearchFile("linux", "amd64", latestStable.String(), info)
	if fileInfo == nil {
		return fmt.Errorf("no fileinfo found")
	}
	err = DLURL(DLURLBase+fileInfo.Filename, fileInfo.Filename, fileInfo.SHA256, fileInfo.Size)
	if err != nil {
		return err
	}
	ts, err := TarStatsFromFile(fileInfo.Filename)
	if err != nil {
		return err
	}
	spew.Dump(ts)
	dstPath := c.String("dstdir")
	err = UnTarFile(fileInfo.Filename, dstPath)
	if err != nil {
		return err
	}
	return nil
}
