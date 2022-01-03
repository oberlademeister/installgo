package main

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
)

func fileSha(name string, expected []byte) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}

	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	got := h.Sum(nil)
	return bytes.Equal(expected, got), nil
}
