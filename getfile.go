package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mitchellh/ioprogress"
	"github.com/sirupsen/logrus"
)

// DLURL downloads a url into a file in local directory
func DLURL(url, name, shastring string, size int64) error {
	l := logrus.WithFields(logrus.Fields{
		"url":       url,
		"name":      name,
		"shastring": shastring,
		"size":      size,
	})
	shaBytes, err := hex.DecodeString(shastring)
	if err != nil {
		return err
	}
	if s, err := os.Stat(name); err == nil && s.Size() == size {
		matches, err := fileSha(name, shaBytes)
		if err == nil && matches {
			// we are done here
			l.Infof("found file with proper size and proper sha")
			return nil
		}
	}

	l.Infof("requesting")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unwanted status code")
	}
	if resp.ContentLength != size {
		return fmt.Errorf("wrong size (expected %d, got %d)", size, resp.ContentLength)
	}
	out, err := os.Create(name)
	l = l.WithFields(logrus.Fields{
		"statuscode":   resp.StatusCode,
		"conentlength": resp.ContentLength,
	})
	if err != nil {
		return err
	}
	l.Infof("downloading")
	progressR := &ioprogress.Reader{
		Reader:       resp.Body,
		Size:         size,
		DrawFunc:     drawProgressLogrus(url),
		DrawInterval: 5 * time.Second,
	}
	written, err := io.Copy(out, progressR)
	if err != nil {
		return err
	}
	if written != size {
		return fmt.Errorf("wrong size written (expected %d, got %d)", size, written)
	}
	match, err := fileSha(name, shaBytes)
	if err != nil {
		return fmt.Errorf("failed to get sha256 from file: %w", err)
	}
	if !match {
		return fmt.Errorf("file downloaded but sha256 does not match")
	}
	return nil
}
