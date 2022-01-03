package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type DLPageResult struct {
	Version string
	Stable  bool
	Files   []*DLPageResultFile
}

type DLPageResultFile struct {
	Filename string
	OS       string
	Arch     string
	Version  string
	SHA256   string
	Size     int64
	Kind     string
}

func RetrievePageResult(url string) ([]*DLPageResult, error) {
	var ret []*DLPageResult
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("unwanted status code or return content-type: %d/%s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling: %w", err)
	}
	resp.Body.Close()
	return ret, nil
}

type Version [3]int

func (v Version) String() string {
	return fmt.Sprintf("go%d.%d.%d", v[0], v[1], v[2])
}

func NewVersion(s string) Version {
	l := logrus.WithField("vstring", s)
	var ret Version
	if !strings.HasPrefix(s, "go") {
		l.Warnf("can't parse version")
		return ret
	}
	s = s[2:]
	vals := strings.Split(s, ".")
	if len(vals) != 3 {
		l.Warnf("can't parse version")
		return ret
	}
	for i := 0; i < 3; i++ {
		var err error
		ret[i], err = strconv.Atoi(vals[i])
		if err != nil {
			l.Warnf("can't parse version")
			return ret
		}
	}
	return ret
}

type VersionList []Version

func (a VersionList) Len() int { return len(a) }
func (a VersionList) Less(i, j int) bool {
	if a[i][0] == a[j][0] {
		if a[i][1] == a[j][1] {
			return a[i][2] < a[j][2]
		}
		return a[i][1] < a[j][1]
	}
	return a[i][0] < a[j][0]
}
func (a VersionList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func GetLatestStableVersion(r []*DLPageResult) (Version, error) {
	var versions VersionList
	for _, v := range r {
		if v.Stable {
			nv := NewVersion(v.Version)
			versions = append(versions, nv)
		}
	}
	sort.Sort(sort.Reverse(versions))
	return versions[0], nil
}

func SearchFile(os, arch, version string, r []*DLPageResult) *DLPageResultFile {
	l := logrus.WithFields(logrus.Fields{
		"os":      os,
		"arch":    arch,
		"version": version,
	})
	for _, e := range r {
		if e.Version == version {
			for _, f := range e.Files {
				if f.Version != version {
					continue
				}
				if f.Arch != arch {
					continue
				}
				if f.OS != os {
					continue
				}
				return f
			}
		}
	}
	l.Warnf("no file found")
	return nil
}
