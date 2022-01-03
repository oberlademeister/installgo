package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func UnTar(dstPath string, r io.Reader) error {
	l := logrus.WithField("dstPath", dstPath)
	l.Info("starting to untar")
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}
		l2 := l.WithFields(logrus.Fields{
			"name": header.Name,
		})
		l2.Info("unpacking tar object")
	}
}

type WalkerFunc func(hdr *tar.Header, tr *tar.Reader) error

func TarWalk(r io.Reader, f WalkerFunc, continueOnWalkErr bool) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}
		walkErr := f(header, tr)
		if walkErr != nil && !continueOnWalkErr {
			return walkErr
		}
	}
}

func GetUntarWalkerFunc(dstPath string) WalkerFunc {
	return func(hdr *tar.Header, tr *tar.Reader) error {
		target := filepath.Join(dstPath, hdr.Name)

		switch hdr.Typeflag {

		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
		return nil
	}
}

type TarStats struct {
	NumFiles      int
	SumBytesFiles int64
	NumDirs       int
}

func (ts *TarStats) Walker(hdr *tar.Header, tr *tar.Reader) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		ts.NumDirs++
	case tar.TypeReg:
		ts.NumFiles++
		ts.SumBytesFiles += hdr.Size
	}
	return nil
}

func TarStatsFromFile(name string) (*TarStats, error) {
	ts := &TarStats{}
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	r := io.Reader(f)
	if strings.HasSuffix(name, "gz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		r = io.Reader(gzr)
	}
	err = TarWalk(r, ts.Walker, false)
	if err != nil {
		return nil, err
	}
	return ts, err
}

func UnTarFile(name, dstPath string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	r := io.Reader(f)
	if strings.HasSuffix(name, "gz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		r = io.Reader(gzr)
	}
	err = TarWalk(r, GetUntarWalkerFunc(dstPath), false)
	if err != nil {
		return err
	}
	return err
}
