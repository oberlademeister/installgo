package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const DLPageInfoURL = "https://go.dev/dl/?mode=json"
const DLURLBase = "https://go.dev/dl/"

func main() {
	app := &cli.App{
		Name:    "installgo",
		Usage:   "install go form main repo into ./go",
		Version: version,
		Action:  run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ddir",
				Aliases: []string{"d"},
				Usage:   "valid path where goinstall will unpack tha tar",
				EnvVars: []string{"GOINSTALLDSTPATH"},
			},
			&cli.StringFlag{
				Name:    "wdir",
				Aliases: []string{"w"},
				Usage:   "valid path where goinstall will store the tar.gz",
				EnvVars: []string{"GOINSTALLTMPPATH"},
			},
			&cli.StringFlag{
				Name:    "os",
				Aliases: []string{"o"},
				Usage:   "operating system determining the proper download",
				EnvVars: []string{"GOINSTALLOS"},
				Value:   runtime.GOOS,
			},
			&cli.StringFlag{
				Name:    "arch",
				Aliases: []string{"a"},
				Usage:   "arch determining the proper downloaod",
				EnvVars: []string{"GOINSTALLOS"},
				Value:   runtime.GOARCH,
			},
			&cli.BoolFlag{
				Name:    "cleanafter",
				Aliases: []string{"c"},
				Usage:   "if set, goinstall will remove the tmp file (the tar.gz)",
				EnvVars: []string{"GOINSTALLCLEANAFTER"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "nounpack",
				Aliases: []string{"n"},
				Usage:   "if set, goinstall will not untar",
				EnvVars: []string{"GOINSTALLNOUNPACK"},
				Value:   false,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	var err error
	wDir := c.String("wdir")
	dDir := c.String("ddir")
	goos := c.String("os")
	goarch := c.String("arch")
	cleanAfter := c.Bool("cleanafter")
	noUnpack := c.Bool("nounpack")

	if wDir == "" {
		wDir, err = ioutil.TempDir("/tmp", "goinstall")
		if err != nil {
			logrus.Fatal(err)
		}
		if cleanAfter {
			defer os.RemoveAll(wDir)
		}
	} else {
		if !filepath.IsAbs(wDir) {
			wDir, err = filepath.Abs(wDir)
			if err != nil {
				return nil
			}
		}
	}
	if dDir == "" && !noUnpack {
		return fmt.Errorf("need a destination directory unloss noUnpack is set")
	}
	dDir, err = filepath.Abs(filepath.Clean(dDir))
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"wDir":       wDir,
		"dDir":       dDir,
		"cleanAfter": cleanAfter,
		"nounpack":   noUnpack,
		"os":         goos,
		"arch":       goarch,
		"version":    version,
		"commit":     commit,
		"date":       date,
	}).Infof("startup")

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
	fileInfo := SearchFile(goos, goarch, latestStable.String(), info)
	if fileInfo == nil {
		return fmt.Errorf("no fileinfo found")
	}

	// download tgz to target

	tgzTarget := filepath.Join(wDir, fileInfo.Filename)
	err = DLURL(DLURLBase+fileInfo.Filename, tgzTarget, fileInfo.SHA256, fileInfo.Size)
	if err != nil {
		return err
	}

	// get tar stats

	stats, err := TarStatsFromFile(tgzTarget)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"numFiles": stats.NumFiles,
		"numDirs":  stats.NumDirs,
		"bytes":    stats.SumBytesFiles,
	}).Info("tgz stats")

	// untar

	l := logrus.WithFields(logrus.Fields{
		"tgz":  tgzTarget,
		"ddir": dDir,
	})
	if noUnpack {
		l.Info("not unpacking")
		return nil
	}
	l.Info("start unpacking")
	err = UnTarFile(fileInfo.Filename, dDir)
	if err != nil {
		return err
	}
	l.Info("unpacking done without error")
	return nil
}
