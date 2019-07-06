package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/prometheus-monitoring/DescribeInstance/lib"
	"github.com/sirupsen/logrus"
)

func writeFile(content []byte, dir string) {
	err := ioutil.WriteFile(dir, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func ensureDir(dir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// New logger
	logrus.SetFormatter(&logrus.TextFormatter{})
	var loglevel = logrus.New()
	loglevel.Out = os.Stdout

	// New targets
	ts := new(lib.Targets)
	var wg sync.WaitGroup

	desDir := "targets/"
	ensureDir(desDir)

	switch arg := os.Args[1]; arg {
	case "all":
		wg.Add(3)
		fallthrough
	case "aws":
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsAWS(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_aws.json"
				loglevel.Info("Write all targets on aws to json file")
				writeFile(content, filedir)
				loglevel.Info("Write targets on aws completed")
			} else {
				loglevel.Error(err)
			}
		}()
		if arg != "all" {
			break
		}
		fallthrough
	case "gcp":
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsGCP(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_gcp.json"
				loglevel.Info("Write all targets on gcp to json file")
				writeFile(content, filedir)
				loglevel.Info("Write targets on gcp completed")
			} else {
				loglevel.Error(err)
			}
		}()
		if arg != "all" {
			break
		}
		fallthrough
	case "vng":
		wg.Add(1)
		go func() {
			defer wg.Done()
			targets, err := ts.GetTargetsVNG(loglevel)
			if err == nil {
				content, _ := json.MarshalIndent(targets, "", "\t")
				filedir := desDir + "targets_vng.json"
				loglevel.Info("Write all targets on datacenter vng to json file")
				writeFile(content, filedir)
				loglevel.Info("Write targets on datacenter vng completed")
			} else {
				loglevel.Error(err)
			}
		}()
	}
	wg.Wait()
}
