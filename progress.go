package main

import "github.com/sirupsen/logrus"

func drawProgressLogrus(url string) func(int64, int64) error {
	return func(current, total int64) error {
		if current < 0 || total < 0 {
			return nil
		}
		l := logrus.WithFields(logrus.Fields{
			"url":     url,
			"current": current,
			"total":   total,
		})
		l.Infof("downloading")
		return nil
	}
}
